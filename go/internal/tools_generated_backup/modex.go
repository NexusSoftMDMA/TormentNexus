package tools

import (
	"context"
)

func HandleHello(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return ok("Hello from Modex!")
}

	return ok("Hello, " + name + "! Welcome to Modex.")
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message argument is required")
}

	return ok("Echo: " + msg)
}