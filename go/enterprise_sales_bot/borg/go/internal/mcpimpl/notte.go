package mcpimpl

import (
	"context"
)

func HandleEcho_notte(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message is required")
}

	return success("Echo: " + msg)
}

func HandleCount_notte(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	count := len(text)
	return ok(map[string]interface{}{"count": count})
}