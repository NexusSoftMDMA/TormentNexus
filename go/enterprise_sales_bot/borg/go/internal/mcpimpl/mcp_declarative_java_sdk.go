package mcpimpl

import (
	"context"
	"fmt"
)

func HandleEcho_mcp_declarative_java_sdk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(fmt.Sprintf("Echo: %s", msg))
}

func HandleHello_mcp_declarative_java_sdk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok(fmt.Sprintf("Hello, %s!", name))
}