package tools

import (
	"context"
)

// HandleMcphubDesktop returns a simple message for Mcphub Desktop.
func HandleMcphubDesktop(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return ok("Hello from Mcphub Desktop MCP server!")
}

	return success("Hello, " + name + "! Welcome to Mcphub Desktop.")
}