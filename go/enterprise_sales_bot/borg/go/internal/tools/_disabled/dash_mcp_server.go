package tools

import (
    "context"
)

func HandleDashHello(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        return err("name is required")
}

    return ok("Hello, " + name + "! Welcome to Dash MCP Server.")
}

func HandleDashStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return ok("Dash MCP Server is running.")
}