package mcpimpl

import (
	"context"
	"fmt"
)

func HandleGreet_decodo_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok(fmt.Sprintf("Hello, %s!", name))
}

func HandlePing_decodo_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}