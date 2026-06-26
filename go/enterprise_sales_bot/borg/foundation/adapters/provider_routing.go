package adapters

import "strings"

type ProviderRouteRequest struct {
	TaskType       string `json:"taskType,omitempty"`
	CostPreference string `json:"costPreference,omitempty"`
	RequireLocal   bool   `json:"requireLocal,omitempty"`
}

type ProviderRoute struct {
	Provider string   `json:"provider"`
	Model    string   `json:"model"`
	Reasons  []string `json:"reasons,omitempty"`
}

func SelectProviderRoute(req ProviderRouteRequest) ProviderRoute {
	status := BuildProviderStatus()
	provider := status.CurrentProvider
	model := status.CurrentModel
	reasons := make([]string, 0, 4)

	if req.RequireLocal && containsString(status.Available, "ollama") {
		provider = "ollama"
		if len(status.OllamaModels) > 0 {
			model = status.OllamaModels[0]
		}
		reasons = append(reasons, "local execution requested and ollama is available")
	}

	switch strings.ToLower(strings.TrimSpace(req.CostPreference)) {
	case "cheap", "low", "budget":
		if containsString(status.Available, "google") {
			provider = "google"
			if model == "" || provider != status.CurrentProvider {
				model = "gemini-2.5-flash"
			}
			reasons = append(reasons, "budget preference favored a lower-cost provider profile")
		} else {
			reasons = append(reasons, "budget preference requested but no alternate provider detected")
		}
	case "high", "quality", "best":
		if containsString(status.Available, "openai") {
			provider = "openai"
			if model == "" || provider != status.CurrentProvider {
				model = "gpt-4o"
			}
			reasons = append(reasons, "quality preference favored the strongest available default route")
		}
	}

	switch strings.ToLower(strings.TrimSpace(req.TaskType)) {
	case "code", "coding", "edit", "refactor":
		reasons = append(reasons, "coding workload detected")
	case "search", "analysis", "research":
		reasons = append(reasons, "analysis workload detected")
	case "local":
		if containsString(status.Available, "ollama") {
			provider = "ollama"
			if len(status.OllamaModels) > 0 {
				model = status.OllamaModels[0]
			}
			reasons = append(reasons, "local-only task type requested")
		}
	}

	if provider == "" {
		provider = "openai"
		reasons = append(reasons, "fallback provider defaulted to openai")
	}
	if model == "" {
		switch provider {
		case "ollama":
			if len(status.OllamaModels) > 0 {
				model = status.OllamaModels[0]
			} else {
				model = "llama3"
			}
		case "google":
			model = "gemini-2.5-flash"
		default:
			model = "gpt-4o"
		}
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "current configured provider and model remained suitable")
	}
	return ProviderRoute{Provider: provider, Model: model, Reasons: reasons}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), target) {
			return true
		}
	}
	return false
}
