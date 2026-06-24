package tools

import (
	"context"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return ok("Hello " + name + " from Xero MCP")
}