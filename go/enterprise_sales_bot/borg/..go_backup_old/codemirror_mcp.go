package tools

import (
	"context"
)

func HandleCodemirror(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Codemirror MCP server is running")
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message is empty")
}

	return ok("Echo: " + msg)
}