package mcpimpl

import (
    "context"
)

func HandleX_mcp_devtools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        return err("name is required")
}

    return success("Tool: " + name)
}