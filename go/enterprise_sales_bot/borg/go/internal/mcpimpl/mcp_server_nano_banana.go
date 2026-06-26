package mcpimpl

import (
    "context"
)

func HandleGreet_mcp_server_nano_banana(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    return ok("Hello, " + name)
}