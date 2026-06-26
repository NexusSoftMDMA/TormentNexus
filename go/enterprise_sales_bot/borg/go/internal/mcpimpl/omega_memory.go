package mcpimpl

import (
	"context"
)

func HandleEcho_omega_memory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	return success("Echo: " + message)
}