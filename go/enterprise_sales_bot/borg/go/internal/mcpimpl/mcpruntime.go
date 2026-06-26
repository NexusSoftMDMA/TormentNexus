package mcpimpl

import (
	"context"
	"time"
)

func HandleEcho_mcpruntime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(msg)
}

func HandleGetTime_mcpruntime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now().Format(time.RFC3339)
	return ok(now)
}