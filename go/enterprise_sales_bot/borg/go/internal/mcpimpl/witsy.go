package mcpimpl

import (
	"context"
	"time"
)

func HandleGreet_witsy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return success("Hello, " + name + "!")
}

func HandleTime_witsy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now().Format(time.RFC3339)
	return ok(now)
}