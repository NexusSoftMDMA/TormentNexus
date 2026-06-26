package mcpimpl

import (
	"context"
)

func HandleEcho_ingero(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(msg)
}