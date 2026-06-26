package tools

import "context"

func HandlePumperlyMcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return ok("Pumperly Mcp ready. Use name parameter for greeting.")
}

	return ok("Hello, " + name + "! Welcome to Pumperly Mcp.")
}