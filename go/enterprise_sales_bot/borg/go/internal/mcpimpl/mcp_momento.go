package mcpimpl

import "context"

func HandleEcho_mcp_momento(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(msg)
}

func HandlePing_mcp_momento(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("pong")
}