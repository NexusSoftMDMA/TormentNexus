//go:build ignore
// +build ignore

package tools

import (
    "context"
    "time"
)

func HandleGetTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    now := time.Now().Format(time.RFC1123)
    return success("Current time: " + now)
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    msg, _ :=getString(args, "message")
    return ok(msg)
}
