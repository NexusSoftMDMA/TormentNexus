package mcpimpl

import (
    "context"
)

func HandleLifeInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return success("Life MCP server is running")
}

func HandleLifePlay(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    action, _ :=getString(args, "action")
    if action == "" {
        return err("action is required")
}

    return ok("Executed action: " + action)
}