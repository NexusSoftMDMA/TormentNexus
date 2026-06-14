package tools

import "context"

func HandlePexbotMcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return success("Hello, " + name + "! Welcome to Pexbot Mcp.")
}