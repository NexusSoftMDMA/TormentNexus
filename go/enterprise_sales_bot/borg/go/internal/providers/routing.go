package providers

type RoutingCandidate struct {
	Provider      string `json:"provider"`
	Configured    bool   `json:"configured"`
	Authenticated bool   `json:"authenticated"`
	Reason        string `json:"reason"`
}

type RoutingTaskSummary struct {
	TaskType   string             `json:"taskType"`
	Strategy   string             `json:"strategy"`
	Candidates []RoutingCandidate `json:"candidates"`
}

type RoutingSummary struct {
	DefaultStrategy string               `json:"defaultStrategy"`
	Tasks           []RoutingTaskSummary `json:"tasks"`
	Limitations     []string             `json:"limitations"`
}

var taskStrategies = map[string]string{
	"coding":     "cheapest",
	"planning":   "best",
	"research":   "best",
	"general":    "round-robin",
	"worker":     "best",
	"supervisor": "best",
}

var taskProviderOrder = map[string][]string{
	"coding":     {"google", "openai", "deepseek", "anthropic", "openrouter", "github-copilot"},
	"planning":   {"anthropic", "openai", "google", "openrouter", "github-copilot", "deepseek"},
	"research":   {"anthropic", "google", "openai", "openrouter", "deepseek", "github-copilot"},
	"general":    {"google", "openai", "anthropic", "deepseek", "openrouter", "github-copilot"},
	"worker":     {"lmstudio", "openrouter", "google", "openai", "deepseek", "anthropic", "github-copilot"},
	"supervisor": {"anthropic", "google", "openai", "openrouter", "github-copilot", "deepseek"},
}

func BuildRoutingSummary(statuses []Status) RoutingSummary {
	statusByProvider := make(map[string]Status, len(statuses))
	for _, status := range statuses {
		statusByProvider[status.Provider] = status
	}

	tasks := make([]RoutingTaskSummary, 0, len(taskStrategies))
	for _, taskType := range []string{"coding", "planning", "research", "general", "worker", "supervisor"} {
		candidates := make([]RoutingCandidate, 0, len(taskProviderOrder[taskType]))
		for index, provider := range taskProviderOrder[taskType] {
			status := statusByProvider[provider]
			reason := "not configured in Go sidecar environment"
			if status.Configured {
				reason = "configured provider in preferred Node routing order"
			} else if index == 0 {
				reason = "preferred provider for this task type when credentials exist"
			}
			candidates = append(candidates, RoutingCandidate{
				Provider:      provider,
				Configured:    status.Configured,
				Authenticated: status.Authenticated,
				Reason:        reason,
			})
		}

		tasks = append(tasks, RoutingTaskSummary{
			TaskType:   taskType,
			Strategy:   taskStrategies[taskType],
			Candidates: candidates,
		})
	}

	return RoutingSummary{
		DefaultStrategy: "best",
		Tasks:           tasks,
		Limitations: []string{
			"Read-only summary only; this does not execute quota-aware model selection.",
			"Ordering reflects current Node task defaults and provider preferences, filtered by configured credentials in the Go sidecar environment.",
			"Live rate limits, quota exhaustion, and fallback events remain Node control-plane responsibilities.",
		},
	}
}
