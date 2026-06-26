package mcpimpl

import (
	"context"
)

// HandleExecuteCommand processes a command for the Mctl Mcp server.
func HandleExecuteCommand_mctl_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("command is required")
}

	return ok("executed: " + command)
}