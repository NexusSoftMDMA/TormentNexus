package tools

import (
	"context"
	"fmt"
)

func HandleHello(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Hello from Lara Mcp")
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(fmt.Sprintf("Echo: %s", msg))
}