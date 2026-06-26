package providers

type CatalogEntry struct {
	Provider       string   `json:"provider"`
	Name           string   `json:"name"`
	AuthMethod     string   `json:"authMethod"`
	DefaultModel   string   `json:"defaultModel"`
	PreferredTasks []string `json:"preferredTasks,omitempty"`
	Executable     bool     `json:"executable"`
	Configured     bool     `json:"configured"`
	Authenticated  bool     `json:"authenticated"`
	EnvVars        []string `json:"envVars,omitempty"`
}

func Catalog(statuses []Status) []CatalogEntry {
	statusByProvider := make(map[string]Status, len(statuses))
	for _, status := range statuses {
		statusByProvider[status.Provider] = status
	}

	entries := []CatalogEntry{
		{
			Provider:       "google",
			Name:           "Google Gemini",
			AuthMethod:     "api_key",
			DefaultModel:   "gemini-2.5-flash",
			PreferredTasks: []string{"coding", "research"},
			Executable:     true,
		},
		{
			Provider:       "anthropic",
			Name:           "Anthropic",
			AuthMethod:     "api_key",
			DefaultModel:   "claude-3-7-sonnet-20250219",
			PreferredTasks: []string{"planning", "research"},
			Executable:     true,
		},
		{
			Provider:       "openai",
			Name:           "OpenAI",
			AuthMethod:     "api_key",
			DefaultModel:   "gpt-4o",
			PreferredTasks: []string{"planning", "coding"},
			Executable:     true,
		},
		{
			Provider:       "deepseek",
			Name:           "DeepSeek",
			AuthMethod:     "api_key",
			DefaultModel:   "deepseek-chat",
			PreferredTasks: []string{"coding", "research"},
			Executable:     true,
		},
		{
			Provider:     "openrouter",
			Name:         "OpenRouter",
			AuthMethod:   "api_key",
			DefaultModel: "openrouter/free",
			Executable:   true,
		},
		{
			Provider:     "github-copilot",
			Name:         "GitHub Copilot",
			AuthMethod:   "pat",
			DefaultModel: "copilot/gpt-4.1",
			Executable:   false,
		},
		{
			Provider:     "google-oauth",
			Name:         "Google OAuth",
			AuthMethod:   "oauth",
			DefaultModel: "google-oauth/gemini",
			Executable:   false,
		},
		{
			Provider:       "lmstudio",
			Name:           "LM Studio",
			AuthMethod:     "none",
			DefaultModel:   "C:/Users/hyper/.lmstudio/models/HauhauCS/Gemma-4-E2B-Uncensored-HauhauCS-Aggressive/Gemma-4-E2B-Uncensored-HauhauCS-Aggressive-Q2_K_P.gguf gemma-4-e2b-uncensored-hauhaucs-aggressive",
			PreferredTasks: []string{"general", "worker"},
			Executable:     true,
		},
		{
			Provider:       "ollama",
			Name:           "Ollama",
			AuthMethod:     "none",
			DefaultModel:   "gemma:2b",
			PreferredTasks: []string{"general"},
			Executable:     true,
		},
	}

	for index, entry := range entries {
		status, ok := statusByProvider[entry.Provider]
		if !ok {
			continue
		}
		entries[index].Configured = status.Configured
		entries[index].Authenticated = status.Authenticated
		entries[index].EnvVars = append([]string(nil), status.EnvVars...)
	}

	return entries
}
