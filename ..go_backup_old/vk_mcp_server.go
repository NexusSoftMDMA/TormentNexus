package tools

import "context"

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("VK MCP Server is running")
}