package mcpimpl

import (
	"context"
)

func HandleGreet_gralio_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok("Hello, " + name + "!")
}

func HandleStatus_gralio_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Gralio Mcp is running")
}