package mcpimpl

import (
	"context"
)

func HandleGreet_maxmsp_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	msg := "Hello, " + name + "! Welcome to Maxmsp MCP Server."
	return ok(msg)
}