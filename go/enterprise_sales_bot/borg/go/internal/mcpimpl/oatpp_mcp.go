package mcpimpl

import "context"

func HandlePing_oatpp_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}

func HandleEcho_oatpp_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return success(msg)
}