package tui

import (
	"strings"
	"testing"

	"github.com/NexusSoftMDMA/TormentNexus/agents"
)

func TestBuildPromptResponseIncludesFoundationRoute(t *testing.T) {
	director := agents.NewDirector(&agents.DefaultProvider{})
	msg, err := buildPromptResponse(director, "Analyze the repository")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(msg.Display, "[Foundation Route]") {
		t.Fatalf("expected foundation route in response: %s", msg.Display)
	}
}

func TestBuildShellProposalIncludesExecutionHint(t *testing.T) {
	director := agents.NewDirector(&agents.DefaultProvider{})
	proposal, err := buildShellProposal(director, "list all go files")
	if err != nil {
		t.Fatal(err)
	}
	if proposal.Command == "" {
		t.Fatal("expected shell command")
	}
	if !strings.Contains(proposal.Explanation, "Route provider execution") {
		t.Fatalf("expected execution hint, got %s", proposal.Explanation)
	}
}
