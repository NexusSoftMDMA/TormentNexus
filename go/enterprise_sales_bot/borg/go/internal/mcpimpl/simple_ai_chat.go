package mcpimpl

import (
    "context"
)

func HandleChat_simple_ai_chat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    message, _ :=getString(args, "message")
    response := "You said: " + message
    return ok(response)
}

func HandlePing_simple_ai_chat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return ok("pong")
}