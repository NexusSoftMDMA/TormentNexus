package mcpimpl

import (
	"context"
	"fmt"
)

func HandlePing_lithtrix_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}

func HandleEcho_lithtrix_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	return success(fmt.Sprintf("Echo: %s", message))
}