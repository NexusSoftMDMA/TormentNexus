package mcpimpl

import "context"

func HandleGloriaMcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Gloria Mcp server is ready")
}