package tools

import (
	"context"
	"time"
)

func HandleInstantTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	t := time.Now().Format(time.RFC3339)
	return ok("Current time: " + t)
}