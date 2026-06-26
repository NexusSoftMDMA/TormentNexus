package orchestration

import "testing"

func TestBuildWebhookPlan(t *testing.T) {
	plan := BuildWebhookPlan("repo_updated", "tormentnexus")
	if len(plan.QueueActions) != 1 || plan.QueueActions[0] != "index_codebase" {
		t.Fatalf("unexpected plan: %#v", plan)
	}
	if plan.ClearLogs {
		t.Fatalf("unexpected clear logs flag: %#v", plan)
	}

	clear := BuildWebhookPlan("clear_logs", "ops")
	if !clear.ClearLogs {
		t.Fatalf("expected clear logs plan: %#v", clear)
	}
}
