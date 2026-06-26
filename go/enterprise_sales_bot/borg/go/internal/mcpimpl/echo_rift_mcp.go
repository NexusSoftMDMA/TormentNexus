package mcpimpl

import (
	"context"
)

func HandleEcho_echo_rift_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return success("Echo: " + msg)
}