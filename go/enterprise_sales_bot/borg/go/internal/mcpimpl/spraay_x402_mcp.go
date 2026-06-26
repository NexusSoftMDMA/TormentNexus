package mcpimpl

import (
	"context"
)

func HandleX_spraay_x402_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	return ok("Echo: " + message)
}