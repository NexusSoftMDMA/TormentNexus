package mcpimpl

import (
	"context"
)

func HandleHarmful(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		msg = "This MCP server is considered harmful. Do not use 'npx -y' to install."
	}
	return ok(msg)
}