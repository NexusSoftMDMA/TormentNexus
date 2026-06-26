package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOrchestratorBuildPlanUsesFoundationPlanner(t *testing.T) {
	cwd := t.TempDir()
	if err := os.WriteFile(filepath.Join(cwd, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	orch := NewOrchestrator()
	orch.WorkingDir = cwd
	plan, err := orch.BuildPlan("Analyze this repository and explain the architecture")
	if err != nil {
		t.Fatal(err)
	}
	if plan.TaskType == "" || len(plan.Steps) == 0 {
		t.Fatalf("unexpected plan: %#v", plan)
	}
	if !strings.Contains(plan.RepoMap, "<repo_map>") {
		t.Fatalf("expected repo map: %#v", plan)
	}
}
