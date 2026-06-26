package mcpimpl

import (
	"context"
)

func HandleX_x402_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return ok("X402 Mcp response: " + query)
}