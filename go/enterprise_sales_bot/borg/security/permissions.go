package security

import (
	"log"
)

// AutonomyLevel defines the freedom of action for native Agent operations.
type AutonomyLevel int

const (
	AutonomyLevelNone    AutonomyLevel = iota // Explicit HITL approval for all commands
	AutonomyLevelRead                         // Can read files without asking
	AutonomyLevelExecute                      // Can run safe builds without asking
	AutonomyLevelGod                          // Total bypass (Dangerous)
)

// PermissionManager replicates the massive TS SandboxService and PermissionManager
type PermissionManager struct {
	Level AutonomyLevel
}

func NewPermissionManager(level AutonomyLevel) *PermissionManager {
	return &PermissionManager{
		Level: level,
	}
}

// RequiresApproval determines if the current Tool/Agent request must block for human input.
func (p *PermissionManager) RequiresApproval(actionType string, resource string) bool {
	if p.Level == AutonomyLevelGod {
		return false
	}

	// Complex heuristics mapping...
	log.Printf("[Security] Evaluating action %s on %s at level %d", actionType, resource, p.Level)
	return true
}

func (p *PermissionManager) InterceptDangerousAction(command string) bool {
	// E.g. native regex checking for rm -rf
	return false
}
