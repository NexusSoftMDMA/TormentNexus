package mcpimpl

import (
	"context"
	"time"
)

func HandleCurrentTime_timergy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now().Format(time.RFC3339)
	return ok("Current time: " + now)
}