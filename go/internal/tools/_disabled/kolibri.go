package tools

import (
	"context"
	"fmt"
)

func HandleGetInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return ok("Kolibri MCP server - no name provided")
}

	return ok(fmt.Sprintf("Hello, %s! Kolibri MCP server is running.", name))
}

func HandleStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Kolibri MCP server status: OK")
}