package mcpimpl

import (
	"context"
)

func HandleHello_template_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
	}
	return success("Hello, " + name + "!")
}

func HandleEcho_template_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message is required")
	}
	return ok("Echo: " + msg)
}