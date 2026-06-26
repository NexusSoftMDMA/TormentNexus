package mcpimpl

import (
	"context"
	"fmt"
)

func HandleHello_zebbern(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	msg := fmt.Sprintf("Hello, %s!", name)
	return ok(msg)
}

func HandleEcho_zebbern(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	return ok(message)
}