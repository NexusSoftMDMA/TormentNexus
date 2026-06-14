package tools

import "context"

func HandleOctagon(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return ok("Hello from Octagon MCP Server, " + name)
}