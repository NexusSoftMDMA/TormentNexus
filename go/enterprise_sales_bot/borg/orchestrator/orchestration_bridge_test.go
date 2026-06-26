package orchestrator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildAutoDriveObjectiveUsesFoundationPlan(t *testing.T) {
	cwd := t.TempDir()
	if err := os.WriteFile(filepath.Join(cwd, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	objective := buildAutoDriveObjective("Refactor the codebase and verify the result.", cwd)
	if !strings.Contains(objective, "Original request") {
		t.Fatalf("expected plan-derived objective, got %q", objective)
	}
}
