package tools

import (
	"context"
	"time"
)

func HandleGetCurrentTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now().Format(time.RFC3339)
	return success(map[string]interface{}{"time": now})
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return success(map[string]interface{}{"echo": msg})
}