package mcpimpl

import (
    "context"
    "fmt"
)

func HandleScout(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        return err("name is required")
}

    return ok(fmt.Sprintf("Scout Mcp found: %s", name))
}