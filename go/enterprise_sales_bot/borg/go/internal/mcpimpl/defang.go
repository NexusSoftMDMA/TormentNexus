package mcpimpl

import (
	"context"
)

func HandleEcho_defang(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	return ok(message)
}