package mcpimpl

import (
	"context"
	"strings"
)

func HandleEcho_tactual(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	return success(message)
}

func HandleUppercase_tactual(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	upper := strings.ToUpper(message)
	return success(upper)
}