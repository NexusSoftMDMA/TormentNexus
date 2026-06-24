package tools

import "context"

func HandleRosMcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return ok("Hello from Ros Mcp!")
}

	return ok("Hello, " + name + " from Ros Mcp!")
}