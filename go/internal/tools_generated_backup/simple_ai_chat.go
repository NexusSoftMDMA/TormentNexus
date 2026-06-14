package tools

import (
    "context"
)

func HandleChat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    message, _ :=getString(args, "message")
    response := "You said: " + message
    return ok(response)
}

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return ok("pong")
}