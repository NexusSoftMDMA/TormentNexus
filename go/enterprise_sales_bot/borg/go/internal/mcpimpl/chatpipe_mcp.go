package mcpimpl

import (
	"context"
	"fmt"
)

func HandleChatpipePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("pong")
}

func HandleChatpipeEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg != "" {
		return success(fmt.Sprintf("Echo: %s", msg))
	}
	return ok("No message provided")
}