package mcpimpl

import (
	"context"
	"time"
)

func HandleTick(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now()
	return success("Current time: " + now.Format(time.RFC3339))
}