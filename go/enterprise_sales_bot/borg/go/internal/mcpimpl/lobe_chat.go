package mcpimpl

import (
	"context"
	"fmt"
	"time"
)

func HandleGetTime_lobe_chat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now().Format(time.RFC3339)
	return ok(fmt.Sprintf("Current time is %s", now))
}

func HandleEcho_lobe_chat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	return ok(fmt.Sprintf("Echo: %s", text))
}