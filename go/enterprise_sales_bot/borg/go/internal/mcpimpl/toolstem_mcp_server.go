package mcpimpl

import (
	"context"
	"fmt"
)

func HandleSayHello_toolstem_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return ok(fmt.Sprintf("Hello, %s!", name))
}