package tools

import "context"

func HandleAxint(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return success("Hello from Axint MCP server!")
}

	return success("Hello, " + name + "!")
}