package tools

import "context"

func HandleGreet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "Agile Team"
	}
	return success("Hello, " + name + "! Welcome to Agile Luminary.")
}

func HandleInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Mcp Agile Luminary: Empowering agile teams with AI insights.")
}