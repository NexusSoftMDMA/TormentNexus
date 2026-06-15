package tools

import (
	"context"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	return ok("Echo: " + message)
}