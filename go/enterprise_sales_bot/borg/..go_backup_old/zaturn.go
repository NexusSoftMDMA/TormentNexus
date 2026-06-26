package tools

import (
    "context"
    "time"
)

func HandleZaturnEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    message, _ :=getString(args, "message")
    return ok("Echo: " + message)
}

func HandleZaturnTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    now := time.Now().Format(time.RFC3339)
    return ok("Current time: " + now)
}