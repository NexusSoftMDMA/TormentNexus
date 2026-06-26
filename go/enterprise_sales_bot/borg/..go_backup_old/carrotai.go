package tools

import (
	"context"
)

func HandleHello(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return success("Hello, " + name + "!")
}

func HandleInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Carrotai MCP server - version 1.0")
}