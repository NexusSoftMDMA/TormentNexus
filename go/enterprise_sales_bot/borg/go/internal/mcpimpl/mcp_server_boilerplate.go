package mcpimpl

import (
	"context"
)

func HandleHello_mcp_server_boilerplate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok("Hello, " + name + "!")
}

func HandlePing_mcp_server_boilerplate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}