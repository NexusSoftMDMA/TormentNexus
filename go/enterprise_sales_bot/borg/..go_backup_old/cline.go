package tools

import (
	"context"
	"net/http"
)

func HandleCline(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	// Example: fetch a URL? No, keep it simple.
	_ = http.DefaultClient
	return ok("You said: " + message)
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "text")
	if msg == "" {
		return err("text is required")
}

	return ok("Echo: " + msg)
}