package mcpimpl

import (
	"context"
)

func HandleGreet_hejdar_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	msg := "Hello, " + name + "! Welcome to Hejdar Mcp."
	return ok(msg)
}