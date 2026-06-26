package mcpimpl

import (
	"context"
)

func HandleGetContext_fulcra_context_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	msg := "Hello, " + name + "! Welcome to Fulcra Context MCP."
	return success(msg)
}