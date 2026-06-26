package mcpimpl

import (
	"context"
	"fmt"
)

func HandleInfo_sweeppea_mcp_info(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "Sweeppea Mcp Info"
	}
	msg := fmt.Sprintf("Hello from %s! This MCP server provides information about awesome MCP servers.", name)
	return ok(msg)
}