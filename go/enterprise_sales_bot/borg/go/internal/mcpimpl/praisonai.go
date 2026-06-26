package mcpimpl

import (
    "context"
    "time"
)

func HandleGetCurrentTime_praisonai(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return success("Current time is " + time.Now().Format(time.RFC3339))
}

func HandleEcho_praisonai(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    msg, _ :=getString(args, "message")
    return success("Echo: " + msg)
}