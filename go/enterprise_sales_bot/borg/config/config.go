package config

import (
	"os"
)

type Config struct {
	Provider string // openai, anthropic, ollama
	Model    string
	APIKey   string
	BaseURL  string
}

func LoadConfig() *Config {
	provider := os.Getenv("SUPERCLI_PROVIDER")
	if provider == "" {
		provider = "openai"
	}
	model := os.Getenv("SUPERCLI_MODEL")
	if model == "" {
		model = "gpt-4o"
	}

	return &Config{
		Provider: provider,
		Model:    model,
		APIKey:   os.Getenv("OPENAI_API_KEY"),
		BaseURL:  os.Getenv("OPENAI_BASE_URL"),
	}
}
