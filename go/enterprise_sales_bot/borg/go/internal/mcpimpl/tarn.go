package mcpimpl

import (
	"context"
	"fmt"
)

func HandleHello_tarn(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	msg := fmt.Sprintf("Hello, %s! Welcome to Tarn MCP server.", name)
	return ok(msg)
}