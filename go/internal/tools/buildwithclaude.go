//go:build ignore
// +build ignore

package tools

import "context"

func HandleListTools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok(map[string]interface{}{
	})

func HandleGetTool(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("missing tool name")
	return ok(map[string]interface{}{
		"description": "A sample MCP tool",
	}),
}
}
}