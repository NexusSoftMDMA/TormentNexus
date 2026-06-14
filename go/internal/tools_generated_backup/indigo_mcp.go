package tools

import (
	"context"
	"net/http"
)

func HandleGreet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok("Hello, " + name + "!")
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message is required")
}

	_ = http.DefaultClient
	return success("You said: " + msg)
}