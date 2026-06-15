package tools

import (
	"context"
	"time"
)

// HandleTime returns the current server time.
func HandleTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	format, _ :=getString(args, "format")
	if format == "" {
		format = time.RFC3339
	}
	now := time.Now().Format(format)
	return ok(now)
}