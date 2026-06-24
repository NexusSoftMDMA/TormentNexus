package tools

import (
	"context"
	"fmt"
	"time"
)

func HandleGetTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now().Format(time.RFC3339)
	return ok(fmt.Sprintf("Current time is %s", now))
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	return ok(fmt.Sprintf("Echo: %s", text))
}