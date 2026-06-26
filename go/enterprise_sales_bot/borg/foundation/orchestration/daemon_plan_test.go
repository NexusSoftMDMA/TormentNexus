package orchestration

import "testing"

func TestBuildDaemonSweepPlan(t *testing.T) {
	plan := BuildDaemonSweepPlan(true, "key", []string{"s1", "s2"}, false)
	if len(plan.SessionChecks) != 2 {
		t.Fatalf("unexpected session checks: %#v", plan)
	}
	if !plan.NeedsReindex {
		t.Fatalf("expected reindex requirement: %#v", plan)
	}
	if len(plan.QueueActions) != 3 {
		t.Fatalf("unexpected queue actions: %#v", plan)
	}

	skipped := BuildDaemonSweepPlan(false, "", nil, true)
	if skipped.SkipReason == "" {
		t.Fatalf("expected skip reason: %#v", skipped)
	}
}
