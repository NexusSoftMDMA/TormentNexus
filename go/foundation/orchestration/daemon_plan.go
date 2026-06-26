package orchestration

import (
	"fmt"
	"strings"
)

type DaemonSweepPlan struct {
	Enabled        bool     `json:"enabled"`
	QueueActions   []string `json:"queueActions,omitempty"`
	SessionChecks  []string `json:"sessionChecks,omitempty"`
	NeedsReindex   bool     `json:"needsReindex"`
	Summary        string   `json:"summary"`
	SkipReason     string   `json:"skipReason,omitempty"`
	TelemetryEvent string   `json:"telemetryEvent"`
}

func BuildDaemonSweepPlan(enabled bool, apiKey string, sessionIDs []string, hasPendingIndex bool) DaemonSweepPlan {
	plan := DaemonSweepPlan{Enabled: enabled, TelemetryEvent: "daemon_swept"}
	if !enabled {
		plan.SkipReason = "daemon disabled in settings"
		plan.Summary = "Daemon sweep skipped because settings disabled the keeper daemon."
		return plan
	}
	if strings.TrimSpace(apiKey) == "" || strings.TrimSpace(apiKey) == "placeholder" {
		plan.SkipReason = "missing api key"
		plan.Summary = "Daemon sweep skipped because no valid supervisor key was configured."
		return plan
	}
	plan.SessionChecks = append(plan.SessionChecks, sessionIDs...)
	for _, sessionID := range sessionIDs {
		plan.QueueActions = append(plan.QueueActions, fmt.Sprintf("check_session:%s", sessionID))
	}
	if !hasPendingIndex {
		plan.NeedsReindex = true
		plan.QueueActions = append(plan.QueueActions, "index_codebase")
	}
	if len(plan.QueueActions) == 0 {
		plan.Summary = "Daemon sweep completed with no actions required."
	} else {
		plan.Summary = fmt.Sprintf("Daemon sweep queued %d action(s).", len(plan.QueueActions))
	}
	return plan
}
