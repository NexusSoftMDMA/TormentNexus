package mcpimpl

import "context"

func HandleHello_stitch_ai_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "Stitch Ai User"
	}
	return ok("Hello, " + name + "!")
}