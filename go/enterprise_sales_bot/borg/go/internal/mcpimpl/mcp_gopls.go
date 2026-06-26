package mcpimpl

import "context"

func HandlePing_mcp_gopls(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}