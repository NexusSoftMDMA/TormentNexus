package tools

import (
    "context"
)

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return ok("pong")
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    msg, _ :=getString(args, "message")
    return ok(msg)
}