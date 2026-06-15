package tools

import "context"

func HandleContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Context7 MCP server is running")
}