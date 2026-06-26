package mcpimpl

import (
	"context"
	"net/http"
)

func HandleGreeting_model_context_protocol_rs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return success("Hello, " + name + "!")
}

func HandleEcho_model_context_protocol_rs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	_ = http.DefaultClient
	return ok("Echo: " + message)
}