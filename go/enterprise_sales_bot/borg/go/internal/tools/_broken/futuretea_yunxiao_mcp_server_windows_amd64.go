package mcpimpl

import "context"

func HandleGreeting_futuretea_yunxiao_mcp_server_windows_amd64(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return ok("Hello, " + name + "!")
}

func HandlePing_futuretea_yunxiao_mcp_server_windows_amd64(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}