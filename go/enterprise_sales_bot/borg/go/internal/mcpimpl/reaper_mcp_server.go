package mcpimpl

import "context"

func HandlePing_reaper_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}

func HandleEcho_reaper_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok("echo: " + msg)
}