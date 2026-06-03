package orchestration

import "strings"

type WebhookPlan struct {
	Type         string   `json:"type"`
	Source       string   `json:"source"`
	Summary      string   `json:"summary"`
	QueueActions []string `json:"queueActions,omitempty"`
	ClearLogs    bool     `json:"clearLogs,omitempty"`
	EmitEvent    string   `json:"emitEvent"`
}

func BuildWebhookPlan(eventType, source string) WebhookPlan {
	cleanType := strings.TrimSpace(eventType)
	cleanSource := strings.TrimSpace(source)
	if cleanSource == "" {
		cleanSource = "unknown"
	}
	plan := WebhookPlan{
		Type:      cleanType,
		Source:    cleanSource,
		EmitEvent: "tormentnexus_signal_received",
		Summary:   "Unhandled signal received; recorded for operator review.",
	}
	switch cleanType {
	case "repo_updated", "reindex_all":
		plan.QueueActions = []string{"index_codebase"}
		plan.Summary = "Repository update detected; schedule codebase re-indexing."
	case "issue_detected":
		plan.QueueActions = []string{"check_issues"}
		plan.Summary = "Issue signal detected; schedule issue review workflow."
	case "clear_logs":
		plan.ClearLogs = true
		plan.Summary = "Administrative log clear signal detected."
	default:
		plan.Summary = "Signal type not mapped to a concrete action; telemetry only."
	}
	return plan
}
