package tools

import (
	"context"
)

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message is required")
}

	return success("Echo: " + msg)
}

func HandleCount(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	count := len(text)
	return ok(map[string]interface{}{"count": count})
}