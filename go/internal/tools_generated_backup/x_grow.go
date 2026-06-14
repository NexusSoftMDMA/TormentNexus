package tools

import (
	"context"
	"fmt"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	return ok(fmt.Sprintf("Received: %s", message))
}