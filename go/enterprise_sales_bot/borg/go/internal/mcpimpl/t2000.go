package mcpimpl

import (
    "context"
    "time"
)

func HandleGetTime_t2000(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    now := time.Now().Format(time.RFC3339)
    return success("Current time: " + now)
}

func HandleEcho_t2000(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    message, _ :=getString(args, "message")
    if message == "" {
        return success("Echo: (empty)")
}

    return success("Echo: " + message)
}