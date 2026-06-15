//go:build ignore
// +build ignore

package tools

import (
	"context"
	"time"
)

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message is required")
}

	return ok("Echo: " + msg)
}

func HandleTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Current time: " + time.Now().Format(time.RFC3339))
}
