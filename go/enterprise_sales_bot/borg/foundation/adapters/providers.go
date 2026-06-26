package adapters

import (
	"fmt"
	"os"
	"strings"

	"github.com/NexusSoftMDMA/TormentNexus/config"
	"github.com/NexusSoftMDMA/TormentNexus/llm"
)

type ProviderStatus struct {
	CurrentProvider string   `json:"currentProvider"`
	CurrentModel    string   `json:"currentModel"`
	BaseURL         string   `json:"baseUrl,omitempty"`
	HasAPIKey       bool     `json:"hasApiKey"`
	Available       []string `json:"available,omitempty"`
	OllamaModels    []string `json:"ollamaModels,omitempty"`
	Warnings        []string `json:"warnings,omitempty"`
}

func BuildProviderStatus() ProviderStatus {
	cfg := config.LoadConfig()
	status := ProviderStatus{
		CurrentProvider: cfg.Provider,
		CurrentModel:    cfg.Model,
		BaseURL:         cfg.BaseURL,
		HasAPIKey:       strings.TrimSpace(cfg.APIKey) != "",
		Available:       detectAvailableProviders(cfg),
	}
	if strings.EqualFold(cfg.Provider, "ollama") || strings.Contains(strings.ToLower(cfg.BaseURL), "11434") {
		client := llm.NewOllamaClient(cfg.BaseURL)
		models, err := client.ListModels()
		if err != nil {
			status.Warnings = append(status.Warnings, fmt.Sprintf("ollama unavailable: %v", err))
		} else {
			status.OllamaModels = make([]string, 0, len(models))
			for _, model := range models {
				status.OllamaModels = append(status.OllamaModels, model.Name)
			}
		}
	}
	return status
}

func BuildProviderContext() string {
	status := BuildProviderStatus()
	parts := []string{"[Provider Adapter]"}
	if status.CurrentProvider != "" {
		parts = append(parts, fmt.Sprintf("Current provider: %s", status.CurrentProvider))
	}
	if status.CurrentModel != "" {
		parts = append(parts, fmt.Sprintf("Current model: %s", status.CurrentModel))
	}
	if len(status.Available) > 0 {
		parts = append(parts, fmt.Sprintf("Available providers: %s", strings.Join(status.Available, ", ")))
	}
	if len(status.OllamaModels) > 0 {
		parts = append(parts, fmt.Sprintf("Ollama models: %s", strings.Join(status.OllamaModels, ", ")))
	}
	if len(status.Warnings) > 0 {
		parts = append(parts, fmt.Sprintf("Provider warnings: %s", strings.Join(status.Warnings, "; ")))
	}
	return strings.Join(parts, "\n")
}

func detectAvailableProviders(cfg *config.Config) []string {
	seen := map[string]struct{}{}
	add := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		seen[name] = struct{}{}
	}
	add(cfg.Provider)
	if strings.TrimSpace(os.Getenv("OPENAI_API_KEY")) != "" {
		add("openai")
	}
	if strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY")) != "" {
		add("anthropic")
	}
	if strings.TrimSpace(os.Getenv("GOOGLE_API_KEY")) != "" {
		add("google")
	}
	if strings.TrimSpace(os.Getenv("OLLAMA_HOST")) != "" || strings.EqualFold(cfg.Provider, "ollama") {
		add("ollama")
	}
	available := make([]string, 0, len(seen))
	for provider := range seen {
		available = append(available, provider)
	}
	sortStrings(available)
	return available
}

func sortStrings(values []string) {
	for i := 0; i < len(values); i++ {
		for j := i + 1; j < len(values); j++ {
			if values[j] < values[i] {
				values[i], values[j] = values[j], values[i]
			}
		}
	}
}
