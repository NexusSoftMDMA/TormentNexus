package mcpimpl

import (
    "context"
    "fmt"
)

func HandleHello_nodit_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        name = "World"
    }
    return ok(fmt.Sprintf("Hello, %s! Welcome to Nodit MCP Server.", name))
}

func HandleEcho_nodit_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    msg, _ :=getString(args, "message")
    if msg == "" {
        return err("message parameter is required")
}

    return success(map[string]interface{}{"echo": msg})
}