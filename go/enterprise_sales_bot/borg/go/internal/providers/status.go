package providers

import "os"

type Status struct {
	Provider      string   `json:"provider"`
	AuthMethod    string   `json:"authMethod"`
	Configured    bool     `json:"configured"`
	Authenticated bool     `json:"authenticated"`
	EnvVar        string   `json:"envVar"`
	EnvVars       []string `json:"envVars,omitempty"`
}

func Snapshot() []Status {
	definitions := []struct {
		provider   string
		authMethod string
		envVars    []string
	}{
		{provider: "openai", authMethod: "api_key", envVars: []string{"OPENAI_API_KEY"}},
		{provider: "anthropic", authMethod: "api_key", envVars: []string{"ANTHROPIC_API_KEY"}},
		{provider: "google", authMethod: "api_key", envVars: []string{"GOOGLE_API_KEY", "GEMINI_API_KEY"}},
		{provider: "google-oauth", authMethod: "oauth", envVars: []string{"GOOGLE_OAUTH_ACCESS_TOKEN"}},
		{provider: "openrouter", authMethod: "api_key", envVars: []string{"OPENROUTER_API_KEY"}},
		{provider: "deepseek", authMethod: "api_key", envVars: []string{"DEEPSEEK_API_KEY"}},
		{provider: "xai", authMethod: "api_key", envVars: []string{"XAI_API_KEY"}},
		{provider: "github-copilot", authMethod: "pat", envVars: []string{"COPILOT_PAT", "GITHUB_TOKEN"}},
	}

	statuses := make([]Status, 0, len(definitions))
	for _, definition := range definitions {
		configured := false
		matchedEnvVar := ""
		for _, envVar := range definition.envVars {
			if os.Getenv(envVar) == "" {
				continue
			}
			configured = true
			matchedEnvVar = envVar
			break
		}
		if matchedEnvVar == "" && len(definition.envVars) > 0 {
			matchedEnvVar = definition.envVars[0]
		}
		statuses = append(statuses, Status{
			Provider:      definition.provider,
			AuthMethod:    definition.authMethod,
			Configured:    configured,
			Authenticated: configured,
			EnvVar:        matchedEnvVar,
			EnvVars:       definition.envVars,
		})
	}
	return statuses
}
