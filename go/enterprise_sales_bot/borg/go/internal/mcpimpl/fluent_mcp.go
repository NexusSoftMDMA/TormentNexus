package mcpimpl

import "context"

func HandleFluent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Fluent MCP server is ready")
}