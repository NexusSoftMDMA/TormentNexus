package mcpimpl

import (
	"context"
)

func HandleGreet_mayamcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	msg := "Hello, " + name + "!"
	return ok(msg)
}

func HandleEcho_mayamcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	return ok("You said: " + message)
}