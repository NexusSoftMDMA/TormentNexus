package security

import (
	"testing"
)

func TestPermissionManager(t *testing.T) {
	// Test God Mode (Bypasses all checks)
	godManager := NewPermissionManager(AutonomyLevelGod)
	if godManager.RequiresApproval("execute", "rm -rf /") {
		t.Error("God autonomy should not require approval")
	}

	// Test Read Mode
	readManager := NewPermissionManager(AutonomyLevelRead)
	if !readManager.RequiresApproval("execute", "dangerous_command") {
		t.Error("Read autonomy must require approval for execute actions")
	}

	// Test intercept heuristic
	if readManager.InterceptDangerousAction("test") {
		t.Error("Dummy command should not be intercepted as dangerous yet")
	}
}
