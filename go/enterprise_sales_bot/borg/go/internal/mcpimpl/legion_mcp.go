package mcpimpl

import (
	"context"
)

// LegioHandler returns a greeting from Legion MCP
func LegioHandler(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return ok("Welcome to Legion MCP!")
}

	return success("Hello " + name + ", Legion MCP is ready!")
}