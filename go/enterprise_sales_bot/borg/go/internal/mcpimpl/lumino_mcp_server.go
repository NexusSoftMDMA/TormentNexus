package mcpimpl

import "context"

func HandleLumino(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok("Hello, " + name + " from Lumino MCP Server!")
}