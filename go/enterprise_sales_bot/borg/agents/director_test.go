package agents

import (
	"context"
	"strings"
	"testing"
)

func TestDirectorInitialization(t *testing.T) {
	provider := &DefaultProvider{}
	director := NewDirector(provider)

	if director.GetName() != "Director" {
		t.Errorf("Expected name 'Director', got '%s'", director.GetName())
	}

	if director.GetRole() != "supervisor" {
		t.Errorf("Expected role 'supervisor', got '%s'", director.GetRole())
	}

	if len(director.History) != 1 {
		t.Errorf("Expected initial prompt history length of 1, got %d", len(director.History))
	}

	if !strings.Contains(director.History[0].Content, "TormentNexus TechLead Director") {
		t.Errorf("System prompt missing core identity")
	}
	if director.HyperAdapter == nil {
		t.Errorf("Expected HyperAdapter to be initialized")
	}
}

func TestDirectorHandleInput(t *testing.T) {
	provider := &DefaultProvider{}
	director := NewDirector(provider)

	resp, err := director.HandleInput(context.Background(), "Run full diagnostic.")
	if err != nil {
		t.Fatalf("HandleInput failed: %v", err)
	}

	if !strings.Contains(resp, "Native Go TormentNexus Director") {
		t.Errorf("Unexpected default response: %s", resp)
	}
	if !strings.Contains(resp, "[Director Plan]") {
		t.Errorf("Expected director plan summary in response: %s", resp)
	}
	if _, ok := director.State["lastPlan"]; !ok {
		t.Errorf("Expected lastPlan state to be recorded")
	}

	// 1 (sys) + 1 (user) + 1 (assistant) = 3 messages in history
	if len(director.History) != 3 {
		t.Errorf("History not appending correctly, expected 3, got %d", len(director.History))
	}
}
