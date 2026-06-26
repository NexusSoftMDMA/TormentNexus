package mcpimpl

import "context"

func HandleKickHealth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("KickJS MCP server is healthy")
}

func HandleKickEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	return success("Echo: " + message)
}