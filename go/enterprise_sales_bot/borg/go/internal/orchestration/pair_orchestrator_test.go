package orchestration

import (
	"testing"
)

func TestPairOrchestratorInitialization(t *testing.T) {
	ce := NewConsensusEngine(nil, nil)
	p := NewPairOrchestrator(ce)
	p.SetupFrontierSquad()
	if len(p.Squad) != 4 {
		t.Errorf("expected 4 squad members, got %d", len(p.Squad))
	}
	planner := p.getMemberName(Planner)
	if planner == "Unknown" {
		t.Errorf("failed to get planner name")
	}
}

func TestPairOrchestratorRoleRotation(t *testing.T) {
	ce := NewConsensusEngine(nil, nil)
	p := NewPairOrchestrator(ce)
	p.SetupFrontierSquad()

	// Capture initial roles
	type memberRole struct {
		name string
		role PairRole
	}
	initial := make([]memberRole, 3)
	for i := 0; i < 3; i++ {
		initial[i] = memberRole{p.Squad[i].Name, p.Squad[i].Role}
	}

	p.RotateRoles()

	// Verify rotation
	if p.Squad[0].Role != initial[1].role || p.Squad[1].Role != initial[2].role || p.Squad[2].Role != initial[0].role {
		t.Errorf("rotation failed. initial: %+v, current: %+v", initial, p.Squad[:3])
	}
}
