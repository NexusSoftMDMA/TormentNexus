package adapters

import (
	"strings"
	"testing"
)

func TestPrepareProviderExecutionBuildsRouteSummary(t *testing.T) {
	setenv(t, "SUPERCLI_PROVIDER", "openai")
	setenv(t, "SUPERCLI_MODEL", "gpt-4o")
	setenv(t, "OPENAI_API_KEY", "test-key")
	result := PrepareProviderExecution(ProviderExecutionRequest{Prompt: "Analyze this repository and explain the architecture.", CostPreference: "budget"})
	if result.TaskType == "" {
		t.Fatal("expected inferred task type")
	}
	if result.Route.Provider == "" || result.Route.Model == "" {
		t.Fatalf("unexpected route result: %#v", result)
	}
	if !strings.Contains(result.ExecutionHint, result.Route.Provider) {
		t.Fatalf("unexpected execution hint: %s", result.ExecutionHint)
	}
	if len(result.SelectionNotes) == 0 {
		t.Fatalf("expected selection notes: %#v", result)
	}
}
