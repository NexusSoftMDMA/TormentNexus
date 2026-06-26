package tools

import (
	"context"
)

func HandleListSessions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Active SSH sessions: none")
}

func HandleExecute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("command argument is required")
}

	return success("Executed: " + command)
}