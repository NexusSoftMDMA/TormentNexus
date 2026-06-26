package tools

import (
	"context"
)

func HandleGetContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	msg := "Hello, " + name + "! Welcome to Fulcra Context MCP."
	return success(msg)
}