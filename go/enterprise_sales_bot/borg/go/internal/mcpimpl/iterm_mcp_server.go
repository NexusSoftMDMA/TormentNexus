package mcpimpl

import (
	"context"
	"fmt"
)

func HandleRunCommand_iterm_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("Command is required")
}

	return ok(fmt.Sprintf("Executed command: %s", command))
}

func HandleListSessions_iterm_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sessions := []string{"session1", "session2"}
	return ok(fmt.Sprintf("Sessions: %v", sessions))
}