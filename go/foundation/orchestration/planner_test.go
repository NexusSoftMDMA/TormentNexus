package orchestration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildPlanIncludesExecutionAndRepoMapWhenRelevant(t *testing.T) {
	cwd := t.TempDir()
	if err := os.WriteFile(filepath.Join(cwd, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, err := BuildPlan(PlanRequest{Prompt: "Analyze this repository and explain the architecture", WorkingDir: cwd, Cost: "budget"})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Execution.Route.Provider == "" {
		t.Fatalf("unexpected execution route: %#v", plan.Execution)
	}
	if !plan.RepoMapIncluded || !strings.Contains(plan.RepoMap, "<repo_map>") {
		t.Fatalf("expected repo map: %#v", plan)
	}
	if len(plan.Steps) == 0 {
		t.Fatalf("expected steps: %#v", plan)
	}
}
