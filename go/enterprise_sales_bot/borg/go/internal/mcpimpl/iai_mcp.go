package mcpimpl

import "context"

func HandleGreet_iai_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return ok("Hello, " + name + "!")
}