package shared

import "os"

type ModelConfig struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Model   string `json:"model"`

	ContextWindow int `json:"context_window"`
}

func NewModelConfig() *ModelConfig {
	return &ModelConfig{
		BaseURL:       getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		APIKey:        getEnv("OPENAI_API_KEY", ""),
		Model:         getEnv("OPENAI_MODEL", "gpt-3.5-turbo"),
		ContextWindow: 200000,
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
