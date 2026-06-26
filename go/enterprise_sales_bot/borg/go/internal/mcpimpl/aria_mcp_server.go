package mcpimpl

import (
	"context"
)

func HandleAria(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	action, _ :=getString(args, "action")
	if action == "" {
		return err("No action provided")
}

	return ok("Action " + action + " performed by Aria MCP Server")
}