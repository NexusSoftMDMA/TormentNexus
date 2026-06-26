package mcpimpl

import (
	"context"
)

func HandleHalSmokeTest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok("Hello " + name + " from MCP server Hal!")
}