package tools

import (
    "context"
    "time"
)

func HandleCurrentTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    currentTime := time.Now().Format("2006-01-02 15:04:05")
    return ok("Current time: " + currentTime)
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    message, _ :=getString(args, "message")
    if message == "" {
        return err("Message cannot be empty")
}

    return ok("Echo: " + message)
}