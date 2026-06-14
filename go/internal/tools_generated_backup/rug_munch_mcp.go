package tools

import (
	"context"
)

func HandleGreeting(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return success("Hello, " + name + "!")
}

func HandleStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Rug Munch Mcp is running")
}