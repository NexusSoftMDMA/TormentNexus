package mcpimpl

import (
	"context"
	"fmt"
)

func HandleGreet_mcp_server_naa(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return ok(fmt.Sprintf("Hello, %s!", name))
}