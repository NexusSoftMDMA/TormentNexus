package tools

import (
	"context"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return ok("X402 Mcp response: " + query)
}