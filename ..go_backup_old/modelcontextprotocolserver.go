package tools

import (
	"context"
	"time"
)

func HandleGetCurrentTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now().Format(time.RFC3339)
	return ok(map[string]interface{}{"time": now})
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	return ok(map[string]interface{}{"echo": message})
}