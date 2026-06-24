package tools

import (
	"context"
)

func HandleFlompt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	return ok("Flompt says: " + message)
}