package mcpimpl

import (
	"context"
)

func HandleGreet_d4vinci(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	greeting := "Hello, " + name + "!"
	return ok(greeting)
}

func HandleEcho_d4vinci(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return success(msg)
}