package pi

import "testing"

func TestDefaultFoundationSpecUsesPiDefaults(t *testing.T) {
	spec := DefaultFoundationSpec()
	if spec.Name != "pi-go-foundation" {
		t.Fatalf("unexpected name: %s", spec.Name)
	}
	if len(spec.Agent.InitialState.Tools) != 7 {
		t.Fatalf("expected 7 builtin tools, got %d", len(spec.Agent.InitialState.Tools))
	}
	if spec.Agent.ToolExecution != ToolExecutionParallel {
		t.Fatalf("expected parallel tool execution, got %s", spec.Agent.ToolExecution)
	}
	if len(spec.RunEventSequence) == 0 || spec.RunEventSequence[0] != EventAgentStart || spec.RunEventSequence[len(spec.RunEventSequence)-1] != EventAgentEnd {
		t.Fatalf("unexpected run event sequence: %#v", spec.RunEventSequence)
	}
}
