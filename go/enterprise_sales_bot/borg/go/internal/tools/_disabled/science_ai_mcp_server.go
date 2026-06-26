package tools

import (
	"context"
)

func HandleAskScience(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return ok("Science AI response to: " + query)
}

func HandleGetVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Science Ai Mcp Server v1.0")
}