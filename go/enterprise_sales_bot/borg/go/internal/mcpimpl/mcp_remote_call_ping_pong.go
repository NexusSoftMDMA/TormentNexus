package mcpimpl

import (
	"context"
)

func HandlePing_mcp_remote_call_ping_pong(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("pong")
}

func HandlePong(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("ping")
}