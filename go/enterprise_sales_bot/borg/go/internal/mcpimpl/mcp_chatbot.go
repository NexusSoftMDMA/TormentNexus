package mcpimpl

import (
	"context"
)

func HandleChat_mcp_chatbot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	return ok("You said: " + message)
}

func HandlePing_mcp_chatbot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}