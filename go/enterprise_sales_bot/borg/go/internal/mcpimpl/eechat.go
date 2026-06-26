package mcpimpl

import (
	"context"
)

func HandleEcho_eechat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok("Echo: " + msg)
}

func HandlePing_eechat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("pong")
}