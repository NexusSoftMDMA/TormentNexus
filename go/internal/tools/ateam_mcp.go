//go:build ignore
// +build ignore

package tools

import (
	"context"
	"time"
)

func HandleGetTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now().String()
	return ok(now)
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message is required")
}

	return ok(msg)
}
