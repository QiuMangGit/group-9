// power-station/cmd/main.go (修改版，集成RAG)
package main

import (
	"context"
	"group-9/config"
	"group-9/llm"
	"group-9/repository"
	"group-9/service/client"
	"group-9/service/connect"
	"group-9/service/rag"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.LoadConfig("../config/config.yaml") // shane: ! 注意修改配置文件
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// ==================== 数据库初始化 ====================
	// 数据库连接配置
	dsn := cfg.Database.DSN
	repo, err := repository.NewRepository(dsn)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// ==================== RAG初始化 ====================
	var ragHandler *rag.RAGHandler
	if cfg.RAG.Enabled {
		log.Println("Initializing RAG system...")
		ragConfig := rag.RAGConfig{
			EmbeddingAPI:   cfg.RAG.EmbeddingAPI,
			EmbeddingModel: cfg.RAG.EmbeddingModel,
			APIKey:         cfg.RAG.APIKey,
			TopK:           cfg.RAG.TopK,
			MinSimilarity:  cfg.RAG.MinSimilarity,
		}

		logger := logrus.New()
		ragHandler = rag.NewRAGHandler(ragConfig, logger)

		// 可以在这里预加载一些知识库内容
		ctx := context.Background()
		sampleKnowledge := `
		我是一个智能语音助手，可以帮助您解答各种问题。
		我具备以下功能：
		1. 语音识别和语音合成
		2. 自然语言理解和生成
		3. 基于知识库的问答
		4. 实时对话交互
		
		常见问题：
		Q: 你是谁？
		A: 我是一个智能语音助手，可以与您进行自然对话，回答您的问题。
		
		Q: 你能做什么？
		A: 我可以进行语音对话，回答问题，提供信息查询服务等。
		
		Q: 如何使用语音功能？
		A: 点击页面上的电话图标即可开始语音通话，直接说话即可。
		`

		if err := ragHandler.LoadDocumentsFromText(ctx, sampleKnowledge, cfg.RAG.ChunkSize); err != nil {
			log.Printf("Failed to load sample knowledge: %v", err)
		} else {
			log.Println("Sample knowledge loaded successfully")
		}
	} else {
		log.Println("RAG system is disabled")
	}

	// 2. 初始化路由
	router := gin.Default()
	// ==================== 静态文件服务 ====================
	// 服务前端文件
	router.Static("/static", "./front-end")
	router.StaticFile("/", "./front-end/index.html")
	router.StaticFile("/rag-admin", "./front-end/rag-admin.html")
	
	// 也可以直接通过以下路径访问：
	// http://localhost:8080/static/rag-admin.html
	// http://localhost:8080/rag-admin
	// 添加CORS中间件
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// 用户认证路由
	userHandler := repository.NewUserHandler(repo)
	userAPI := router.Group("/api/users")
	{
		userAPI.POST("/register", userHandler.Register)
		userAPI.POST("/login", userHandler.Login)
		userAPI.POST("/sendEmailCode", userHandler.SendEmailCode)
		userAPI.GET("/getDetail", userHandler.GetDetail)
		userAPI.POST("/submitDetail", userHandler.SubmitDetail)
	}

	// 机器人相关API路由
	robotHandler := repository.NewRobotHandler(repo)
	robotAPI := router.Group("/api/robots")
	{
		robotAPI.POST("/addAssistant", robotHandler.AddAssistant)
		robotAPI.POST("/deleteAssistant", robotHandler.DeleteAssistant)
		robotAPI.POST("/updateAssistant", robotHandler.UpdateAssistant)
		robotAPI.POST("/getAssistant", robotHandler.GetAssistant)
	}

	// 语音对话相关API路由
	voiceConversationHandler := repository.NewConversationHandler(repo)
	voiceConversationAPI := router.Group("/api/chat")
	{
		voiceConversationAPI.POST("/deleteVoiceChatContent", voiceConversationHandler.DeleteVoiceConversation)
		voiceConversationAPI.POST("/selectVoiceChatContent", voiceConversationHandler.GetVoiceConversation)
	}

	// ==================== RAG相关API路由 ====================
	if ragHandler != nil {
		ragAPI := router.Group("/api/rag")
		{
			// 添加知识库内容
			ragAPI.POST("/addKnowledge", func(c *gin.Context) {
				var req struct {
					Text string `json:"text" binding:"required"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				if err := ragHandler.LoadDocumentsFromText(c.Request.Context(), req.Text, cfg.RAG.ChunkSize); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{"message": "Knowledge added successfully"})
			})

			// 获取RAG统计信息
			ragAPI.GET("/stats", func(c *gin.Context) {
				stats := ragHandler.GetDatabaseStats()
				c.JSON(http.StatusOK, stats)
			})

			// 清空知识库
			ragAPI.POST("/clear", func(c *gin.Context) {
				ragHandler.ClearDatabase()
				c.JSON(http.StatusOK, gin.H{"message": "Knowledge base cleared"})
			})

			// 测试检索功能
			ragAPI.POST("/search", func(c *gin.Context) {
				var req struct {
					Query         string  `json:"query" binding:"required"`
					TopK          int     `json:"top_k"`
					MinSimilarity float64 `json:"min_similarity"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				if req.TopK <= 0 {
					req.TopK = cfg.RAG.TopK
				}
				if req.MinSimilarity <= 0 {
					req.MinSimilarity = cfg.RAG.MinSimilarity
				}

				results, err := ragHandler.SearchSimilar(c.Request.Context(), req.Query, req.TopK, req.MinSimilarity)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{"results": results})
			})
		}
	}

	asrOption := &client.ASROption{
		Provider:   cfg.ASR.Provider,
		Language:   cfg.ASR.Language,
		SampleRate: cfg.ASR.SampleRate,
		AppID:      cfg.ASR.AppID,
		SecretID:   cfg.ASR.SecretID,
		SecretKey:  cfg.ASR.SecretKey,
		Endpoint:   cfg.ASR.Endpoint,
		ModelType:  cfg.ASR.ModelType,
	}

	ttsOption := &client.TTSOption{
		Provider:   cfg.TTS.Provider,
		Samplerate: cfg.TTS.SampleRate,
		Speaker:    cfg.TTS.Speaker,
		Speed:      cfg.TTS.Speed,
		Volume:     cfg.TTS.Volume,
		AppID:      cfg.TTS.AppID,
		SecretID:   cfg.TTS.SecretID,
		SecretKey:  cfg.TTS.SecretKey,
		Codec:      cfg.TTS.Codec,
		Endpoint:   cfg.TTS.Endpoint,
	}

	ctx := context.Background()
	logger := logrus.New()

	// 创建LLM处理器时传入RAG处理器
	llmHandler := llm.NewLLMHandler(ctx, cfg.LLM.APIKey, cfg.LLM.URL, cfg.LLM.SystemPrompt, logger, ragHandler)

	r := gin.Default()

	// shane: 后端建立连接
	backendServer := connect.NewBackendServer(cfg.Backend.URL)
	backendConn, err := backendServer.Connect(cfg.Backend.CallType)
	if err != nil {
		log.Fatalf("Unable to connect to backend: %v", err)
	} else {
		log.Println("Connected to backend successfully!")
	}

	// shane: 前端建立连接
	frontendServer := connect.NewFrontendServer(llmHandler, backendConn, backendServer, cfg.Audio.Codec, asrOption, ttsOption, repo)
	frontendServer.Start(r, cfg.Server.Port)

	// 启动服务器
	router.Run(":8080")

	select {}
}
