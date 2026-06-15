package tools

import (
	"context"
	"strings"
)

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	return success(message)
}

func HandleUppercase(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	upper := strings.ToUpper(message)
	return success(upper)
}