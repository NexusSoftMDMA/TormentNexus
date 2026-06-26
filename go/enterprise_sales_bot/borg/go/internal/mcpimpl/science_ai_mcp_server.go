package mcpimpl

import (
	"context"
)

func HandleAskScience(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return ok("Science AI response to: " + query)
}

func HandleGetVersion_science_ai_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Science Ai Mcp Server v1.0")
}