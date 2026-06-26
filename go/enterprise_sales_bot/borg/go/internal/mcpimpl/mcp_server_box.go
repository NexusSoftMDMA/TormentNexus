package mcpimpl

import (
    "context"
)

func HandleBox(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    content, _ :=getString(args, "content")
    return ok(content)
}

func HandlePing_mcp_server_box(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return ok("pong")
}