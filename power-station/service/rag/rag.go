// power-station/service/rag/rag.go
package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// RAGHandler 处理RAG相关的功能
type RAGHandler struct {
	client       *http.Client
	embeddingAPI string
	apiKey       string
	logger       *logrus.Logger
	vectorDB     *VectorDB
}

// VectorDocument 向量文档结构
type VectorDocument struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Vector    []float64              `json:"vector"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

// VectorDB 简单的内存向量数据库
type VectorDB struct {
	documents map[string]*VectorDocument
	index     []*VectorDocument // 用于快速检索
}

// EmbeddingRequest 请求结构
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// EmbeddingResponse 响应结构
type EmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Index     int       `json:"index"`
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// SimilarityResult 相似度搜索结果
type SimilarityResult struct {
	Document   *VectorDocument `json:"document"`
	Similarity float64         `json:"similarity"`
}

// RAGConfig RAG配置
type RAGConfig struct {
	EmbeddingAPI   string  `yaml:"embedding_api"`
	EmbeddingModel string  `yaml:"embedding_model"`
	APIKey         string  `yaml:"api_key"`
	TopK           int     `yaml:"top_k"`
	MinSimilarity  float64 `yaml:"min_similarity"`
}

// NewRAGHandler 创建新的RAG处理器
func NewRAGHandler(config RAGConfig, logger *logrus.Logger) *RAGHandler {
	vectorDB := &VectorDB{
		documents: make(map[string]*VectorDocument),
		index:     make([]*VectorDocument, 0),
	}

	return &RAGHandler{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		embeddingAPI: config.EmbeddingAPI,
		apiKey:       config.APIKey,
		logger:       logger,
		vectorDB:     vectorDB,
	}
}

// GetEmbedding 调用embedding API获取向量
func (r *RAGHandler) GetEmbedding(ctx context.Context, texts []string, model string) ([][]float64, error) {
	if model == "" {
		model = "text-embedding-ada-002"
	}

	reqBody := EmbeddingRequest{
		Model: model,
		Input: texts,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", r.embeddingAPI, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.apiKey)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var embeddingResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	embeddings := make([][]float64, len(embeddingResp.Data))
	for _, data := range embeddingResp.Data {
		embeddings[data.Index] = data.Embedding
	}

	r.logger.WithFields(logrus.Fields{
		"texts_count":      len(texts),
		"embeddings_count": len(embeddings),
		"model":            model,
	}).Debug("Successfully got embeddings")

	return embeddings, nil
}

// AddDocuments 添加文档到向量数据库
func (r *RAGHandler) AddDocuments(ctx context.Context, contents []string, metadata []map[string]interface{}) error {
	if len(contents) != len(metadata) {
		return fmt.Errorf("contents and metadata length mismatch")
	}

	// 获取向量
	embeddings, err := r.GetEmbedding(ctx, contents, "text-embedding-ada-002")
	if err != nil {
		return fmt.Errorf("get embeddings failed: %w", err)
	}

	// 添加到向量数据库
	for i, content := range contents {
		doc := &VectorDocument{
			ID:        generateID(),
			Content:   content,
			Vector:    embeddings[i],
			Metadata:  metadata[i],
			CreatedAt: time.Now(),
		}

		r.vectorDB.documents[doc.ID] = doc
		r.vectorDB.index = append(r.vectorDB.index, doc)
	}

	r.logger.WithField("documents_added", len(contents)).Info("Added documents to vector database")
	return nil
}

// SearchSimilar 搜索相似文档
func (r *RAGHandler) SearchSimilar(ctx context.Context, query string, topK int, minSimilarity float64) ([]*SimilarityResult, error) {
	// 获取查询向量
	queryEmbeddings, err := r.GetEmbedding(ctx, []string{query}, "text-embedding-ada-002")
	if err != nil {
		return nil, fmt.Errorf("get query embedding failed: %w", err)
	}

	if len(queryEmbeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned for query")
	}

	queryVector := queryEmbeddings[0]

	// 计算相似度
	similarities := make([]*SimilarityResult, 0)
	for _, doc := range r.vectorDB.index {
		similarity := cosineSimilarity(queryVector, doc.Vector)
		if similarity >= minSimilarity {
			similarities = append(similarities, &SimilarityResult{
				Document:   doc,
				Similarity: similarity,
			})
		}
	}

	// 按相似度排序
	sort.Slice(similarities, func(i, j int) bool {
		return similarities[i].Similarity > similarities[j].Similarity
	})

	// 返回topK结果
	if topK > 0 && len(similarities) > topK {
		similarities = similarities[:topK]
	}

	r.logger.WithFields(logrus.Fields{
		"query":         query,
		"results_count": len(similarities),
		"top_k":         topK,
	}).Debug("Search completed")

	return similarities, nil
}

// RetrieveContext 检索相关上下文
func (r *RAGHandler) RetrieveContext(ctx context.Context, query string, topK int, minSimilarity float64) (string, error) {
	if topK <= 0 {
		topK = 3
	}
	if minSimilarity <= 0 {
		minSimilarity = 0.7
	}

	results, err := r.SearchSimilar(ctx, query, topK, minSimilarity)
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		r.logger.Debug("No relevant context found")
		return "", nil
	}

	// 构建上下文字符串
	var contextParts []string
	for i, result := range results {
		contextParts = append(contextParts, fmt.Sprintf("参考信息%d (相似度: %.3f):\n%s",
			i+1, result.Similarity, result.Document.Content))
	}

	context := strings.Join(contextParts, "\n\n")

	r.logger.WithFields(logrus.Fields{
		"query":          query,
		"context_parts":  len(contextParts),
		"context_length": len(context),
	}).Debug("Retrieved context")

	return context, nil
}

// LoadDocumentsFromText 从文本文件加载文档（分块处理）
func (r *RAGHandler) LoadDocumentsFromText(ctx context.Context, text string, chunkSize int) error {
	if chunkSize <= 0 {
		chunkSize = 1000 // 默认每块1000字符
	}

	chunks := splitTextIntoChunks(text, chunkSize)
	contents := make([]string, len(chunks))
	metadata := make([]map[string]interface{}, len(chunks))

	for i, chunk := range chunks {
		contents[i] = chunk
		metadata[i] = map[string]interface{}{
			"chunk_index": i,
			"chunk_size":  len(chunk),
			"source":      "text_input",
		}
	}

	return r.AddDocuments(ctx, contents, metadata)
}

// GetDatabaseStats 获取数据库统计信息
func (r *RAGHandler) GetDatabaseStats() map[string]interface{} {
	return map[string]interface{}{
		"total_documents": len(r.vectorDB.documents),
		"index_size":      len(r.vectorDB.index),
	}
}

// ClearDatabase 清空数据库
func (r *RAGHandler) ClearDatabase() {
	r.vectorDB.documents = make(map[string]*VectorDocument)
	r.vectorDB.index = make([]*VectorDocument, 0)
	r.logger.Info("Vector database cleared")
}

// cosineSimilarity 计算余弦相似度
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0.0 || normB == 0.0 {
		return 0.0
	}

	return dotProduct / (sqrt(normA) * sqrt(normB))
}

// sqrt 计算平方根
func sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}

	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// splitTextIntoChunks 将文本分割成块
func splitTextIntoChunks(text string, chunkSize int) []string {
	var chunks []string

	// 按句子分割
	sentences := strings.Split(text, "。")

	var currentChunk strings.Builder

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		// 如果当前块加上新句子超过了限制，就保存当前块并开始新块
		if currentChunk.Len() > 0 && currentChunk.Len()+len(sentence) > chunkSize {
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString("。")
		}
		currentChunk.WriteString(sentence)
	}

	// 添加最后一块
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}

// generateID 生成唯一ID
func generateID() string {
	return fmt.Sprintf("doc_%d", time.Now().UnixNano())
}
