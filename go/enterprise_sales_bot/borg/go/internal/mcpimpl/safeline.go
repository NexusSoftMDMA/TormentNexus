package mcpimpl

import (
	"context"
)

func HandleStatus_safeline(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Safeline MCP server is running")
}

func HandleInfo_safeline(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name parameter required")
}

	return ok("Hello, " + name)
}