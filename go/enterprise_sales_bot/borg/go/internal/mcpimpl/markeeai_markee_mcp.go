package mcpimpl

import (
	"context"
)

func HandleGreet_markeeai_markee_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return ok("Hello, " + name + "!")
}

func HandleEcho_markeeai_markee_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	return success("You said: " + message)
}