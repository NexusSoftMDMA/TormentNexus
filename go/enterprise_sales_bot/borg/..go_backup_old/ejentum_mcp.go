package tools

import (
	"context"
)

func HandleEjentum(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return success("Hello, " + name + " from Ejentum MCP server!")
}