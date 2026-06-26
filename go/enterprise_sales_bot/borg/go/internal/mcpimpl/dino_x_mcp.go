package mcpimpl

import "context"

func HandleX_dino_x_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return success("Hello, " + name + " from Dino X MCP!")
}