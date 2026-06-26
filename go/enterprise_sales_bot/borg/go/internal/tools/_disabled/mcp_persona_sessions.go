package tools

import "context"

func HandleCreateSession(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	persona, _ :=getString(args, "persona")
	// Simulate session creation
	sessionID := "session_" + name + "_" + persona
	return ok("Created session " + sessionID)
}

func HandleGetSession(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sessionID, _ :=getString(args, "session_id")
	if sessionID == "" {
		return err("session_id is required")
}

	// Simulate session retrieval
	return success("Session " + sessionID + " is active")
}