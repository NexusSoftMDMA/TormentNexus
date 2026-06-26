package mcpimpl

import (
    "context"
    "fmt"
)

func HandleGreet_context_apps_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    msg := fmt.Sprintf("Hello, %s!", name)
    return ok(msg)
}

func HandleEcho_context_apps_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    text, _ :=getString(args, "text")
    if text == "" {
        return err("text is required")
}

    return success(text)
}