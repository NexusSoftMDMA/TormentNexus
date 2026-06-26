package mcpimpl

import (
	"context"
	"fmt"
)

func HandleX_x_grow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	return ok(fmt.Sprintf("Received: %s", message))
}