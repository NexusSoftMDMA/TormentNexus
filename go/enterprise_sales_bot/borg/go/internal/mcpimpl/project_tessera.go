package mcpimpl

import (
    "context"
)

func HandlePing_project_tessera(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return ok("pong")
}

func HandleEcho_project_tessera(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    msg, _ :=getString(args, "message")
    return ok(msg)
}