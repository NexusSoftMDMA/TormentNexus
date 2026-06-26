package db

import (
	"testing"
)

func TestLeadStateEnums(t *testing.T) {
	states := []LeadState{
		StateDiscovered,
		StateResearched,
		StateOutreachSent,
		StateEngaged,
		StateNegotiating,
		StateClosedWon,
		StateClosedLost,
	}

	expected := 7
	if len(states) != expected {
		t.Errorf("Expected %d states, got %d", expected, len(states))
	}
}
