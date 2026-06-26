package mcpimpl

import (
	"context"
	"fmt"
	"time"
)

func HandleEcho_quantoracle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message is required")
}

	return ok(fmt.Sprintf("Echo: %s", msg))
}

func HandleTime_quantoracle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	format, _ :=getString(args, "format")
	if format == "" {
		format = time.RFC3339
	}
	return ok(fmt.Sprintf("Current time: %s", time.Now().Format(format)))
}