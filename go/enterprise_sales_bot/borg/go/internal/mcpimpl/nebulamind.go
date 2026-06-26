package mcpimpl

import (
	"context"
)

func HandleInfo_nebulamind(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "Nebulamind"
	}
	return ok("Hello from " + name + " MCP server!")
}

func HandleEcho_nebulamind(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message is required")
}

	return success("Echo: " + msg)
}