package mcpimpl

import "context"

func HandleNocturnusai(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Nocturnusai MCP server is running")
}