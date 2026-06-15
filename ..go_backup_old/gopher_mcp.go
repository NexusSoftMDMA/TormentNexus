package tools

import (
    "context"
    "time"
)

func HandleGetCurrentTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    format, _ :=getString(args, "format")
    if format == "" {
        format = time.RFC3339
    }
    now := time.Now().Format(format)
    return success(now)
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    message, _ :=getString(args, "message")
    if message == "" {
        return err("message is required")
}

    return ok(message)
}