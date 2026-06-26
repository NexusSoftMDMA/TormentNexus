package mcpimpl

import "context"

func HandleReexpress(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	reexpressed := "Re-expressed: " + message
	return ok(reexpressed)
}