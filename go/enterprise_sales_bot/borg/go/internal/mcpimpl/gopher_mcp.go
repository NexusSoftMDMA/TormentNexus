package mcpimpl

import (
    "context"
    "time"
)

func HandleGetCurrentTime_gopher_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    format, _ :=getString(args, "format")
    if format == "" {
        format = time.RFC3339
    }
    now := time.Now().Format(format)
    return success(now)
}

func HandleEcho_gopher_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    message, _ :=getString(args, "message")
    if message == "" {
        return err("message is required")
}

    return ok(message)
}