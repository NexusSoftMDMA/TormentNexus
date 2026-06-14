package tools

import (
	"context"
)

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	return success(message)
}

func HandleInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Turbomcp MCP server v1.0")
}