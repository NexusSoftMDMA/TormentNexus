package tools

import (
	"context"
)

func HandleProdeMc(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
	}
	return success("Hello, " + name + "!")
}

func HandleProdeMcInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Prode Mc MCP server version 1.0.0")
}