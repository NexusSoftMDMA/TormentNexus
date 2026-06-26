package mcpimpl

import (
	"context"
)

func HandleEcho_genkit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message is required")
}

	return success("Echo: " + msg)
}