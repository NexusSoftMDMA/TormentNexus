package tools

import "context"

func HandleGetServerInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Promptarchitect MCP Server v1.0")
}