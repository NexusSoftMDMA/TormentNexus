package mcpimpl

import (
	"context"
)

func HandleEcho_turbomcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	return success(message)
}

func HandleInfo_turbomcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Turbomcp MCP server v1.0")
}