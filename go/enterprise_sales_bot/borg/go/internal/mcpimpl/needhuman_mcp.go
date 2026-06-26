package mcpimpl

import "context"

func HandleHello_needhuman_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok("Hello from Needhuman Mcp, " + name + "!")
}