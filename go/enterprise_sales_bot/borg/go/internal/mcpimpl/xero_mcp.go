package mcpimpl

import (
	"context"
)

func HandleX_xero_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return ok("Hello " + name + " from Xero MCP")
}