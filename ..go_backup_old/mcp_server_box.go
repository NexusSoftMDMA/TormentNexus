package tools

import (
    "context"
)

func HandleBox(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    content, _ :=getString(args, "content")
    return ok(content)
}

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return ok("pong")
}