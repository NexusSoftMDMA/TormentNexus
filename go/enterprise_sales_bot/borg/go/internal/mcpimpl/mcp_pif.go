package mcpimpl

import (
	"context"
	"fmt"
)

func HandlePing_mcp_pif(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}

func HandleEcho_mcp_pif(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message is required")
	}
	return ok(fmt.Sprintf("echo: %s", msg))
}