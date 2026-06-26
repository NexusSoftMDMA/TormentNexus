package mcpimpl

import (
	"context"
	"fmt"
)

func HandleVimoInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	msg := fmt.Sprintf("Hello, %s! This is Vimo MCP Server.", name)
	return success(msg)
}