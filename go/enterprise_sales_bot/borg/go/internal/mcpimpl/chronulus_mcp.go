package mcpimpl

import (
	"context"
	"time"
)

func HandleCurrentTime_chronulus_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	format, _ :=getString(args, "format")
	if format == "" {
		format = time.RFC3339
	}
	now := time.Now()
	return ok(now.Format(format))
}