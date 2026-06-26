package mcpimpl

import (
	"context"
)

func HandleGreeting_rug_munch_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return success("Hello, " + name + "!")
}

func HandleStatus_rug_munch_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Rug Munch Mcp is running")
}