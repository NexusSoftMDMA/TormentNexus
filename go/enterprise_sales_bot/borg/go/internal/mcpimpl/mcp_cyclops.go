package mcpimpl

import "context"

func HandleGreet_mcp_cyclops(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return success("Hello from Cyclops, " + name + "!")
}