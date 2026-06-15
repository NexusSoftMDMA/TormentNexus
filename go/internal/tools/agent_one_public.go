//go:build ignore
// +build ignore

package tools

import "context"

func HandleGetInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Agent One Public MCP server is running")
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		msg = "Echo: no message provided"
	}
	return success(msg)
}
