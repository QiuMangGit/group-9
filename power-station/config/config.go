// power-station/config/config.go (修改版)
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Backend  BackendConfig  `yaml:"backend"`
	Audio    AudioConfig    `yaml:"audio"`
	ASR      ASRConfig      `yaml:"asr"`
	TTS      TTSConfig      `yaml:"tts"`
	LLM      LLMConfig      `yaml:"llm"`
	RAG      RAGConfig      `yaml:"rag"`      // 新增RAG配置
	VAD      VADConfig      `yaml:"vad"`
	Call     CallConfig     `yaml:"call"`
	WebHook  WebHookConfig  `yaml:"webhook"`
	EOU      EOUConfig      `yaml:"eou"`
	Database DatabaseConfig `yaml:"database"`
}

type DatabaseConfig struct {
	DSN string `yaml:"dsn"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type BackendConfig struct {
	URL      string `yaml:"url"`
	CallType string `yaml:"call_type"`
}

type AudioConfig struct {
	Codec string `yaml:"codec"`
}

type ASRConfig struct {
	Provider   string `yaml:"provider"`
	Language   string `yaml:"language"`
	SampleRate uint32 `yaml:"sample_rate"`
	AppID      string `yaml:"app_id"`
	SecretID   string `yaml:"secret_id"`
	SecretKey  string `yaml:"secret_key"`
	Endpoint   string `yaml:"endpoint"`
	ModelType  string `yaml:"model_type"`
}

type TTSConfig struct {
	Provider        string  `yaml:"provider"`
	SampleRate      int32   `yaml:"sample_rate"`
	Speaker         string  `yaml:"speaker"`
	Speed           float32 `yaml:"speed"`
	Volume          int32   `yaml:"volume"`
	EmotionCategory string  `yaml:"emotion"`
	AppID           string  `yaml:"app_id"`
	SecretID        string  `yaml:"secret_id"`
	SecretKey       string  `yaml:"secret_key"`
	Codec           string  `yaml:"codec"`
	Endpoint        string  `yaml:"endpoint"`
}

type LLMConfig struct {
	APIKey       string `yaml:"api_key"`
	Model        string `yaml:"model"`
	URL          string `yaml:"url"`
	SystemPrompt string `yaml:"system_prompt"`
}

// 新增RAG配置结构
type RAGConfig struct {
	Enabled        bool    `yaml:"enabled"`         // 是否启用RAG
	EmbeddingAPI   string  `yaml:"embedding_api"`   // Embedding API地址
	EmbeddingModel string  `yaml:"embedding_model"` // Embedding模型名称
	APIKey         string  `yaml:"api_key"`         // API密钥
	TopK           int     `yaml:"top_k"`           // 检索top-k结果
	MinSimilarity  float64 `yaml:"min_similarity"`  // 最小相似度阈值
	ChunkSize      int     `yaml:"chunk_size"`      // 文本分块大小
}

type VADConfig struct {
	Model     string `yaml:"model"`
	Endpoint  string `yaml:"endpoint"`
	SecretKey string `yaml:"secret_key"`
}

type CallConfig struct {
	BreakOnVAD bool   `yaml:"break_on_vad"`
	WithSIP    bool   `yaml:"with_sip"`
	Record     bool   `yaml:"record"`
	Caller     string `yaml:"caller"`
	Callee     string `yaml:"callee"`
}

type WebHookConfig struct {
	Addr   string `yaml:"addr"`
	Prefix string `yaml:"prefix"`
}

type EOUConfig struct {
	Type     string `yaml:"type"`
	Endpoint string `yaml:"endpoint"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &config, nil
}