package mcpimpl

import (
	"context"
)

func HandleGreet_lokka(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return ok("Hello, world!")
}

	return ok("Hello, " + name + "!")
}

func HandleEcho_lokka(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("no message provided")
}

	return ok("Echo: " + msg)
}