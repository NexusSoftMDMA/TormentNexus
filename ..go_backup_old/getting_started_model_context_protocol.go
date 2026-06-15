package tools

import "context"

func HandleGettingStarted(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return ok("Hello! Welcome to Getting Started with MCP. This course teaches C# development.")
}

	return ok("Hello " + name + "! Welcome to Getting Started with MCP. This course teaches C# development.")
}