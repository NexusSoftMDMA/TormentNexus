package tools

import (
	"context"
)

func HandleGetNei(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok("Hello, " + name + " from NEI MCP server")
}

func HandleListNei(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("List of NEI items: item1, item2, item3")
}