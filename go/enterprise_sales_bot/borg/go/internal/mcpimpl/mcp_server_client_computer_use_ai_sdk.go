package mcpimpl

import "context"

func HandlePing_mcp_server_client_computer_use_ai_sdk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("pong")
}

func HandleEcho_mcp_server_client_computer_use_ai_sdk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(msg)
}