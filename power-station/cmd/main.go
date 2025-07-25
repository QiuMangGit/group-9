package main

import (
	"context"
	"group-9/config"
	"group-9/llm"
	"group-9/repository"
	"group-9/service/client"
	"group-9/service/connect"
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

	// 2. 初始化路由
	router := gin.Default()
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
		userAPI.POST("/sendEmailCode",userHandler.SendEmailCode)
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
	llm := llm.NewLLMHandler(ctx, cfg.LLM.APIKey, cfg.LLM.URL, cfg.LLM.SystemPrompt, logger)

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
	frontendServer := connect.NewFrontendServer(llm, backendConn, backendServer, cfg.Audio.Codec, asrOption, ttsOption,repo)
	frontendServer.Start(r, cfg.Server.Port)

	// 启动服务器
	router.Run(":8080")

	select {}

}
