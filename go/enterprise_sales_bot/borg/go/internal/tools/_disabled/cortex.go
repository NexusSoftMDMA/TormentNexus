package tools

import "context"

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name != "" {
		return ok("Hello, " + name + " from Cortex MCP server")
}

	return ok("Cortex MCP server ready")
}