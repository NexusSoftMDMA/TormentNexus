package mcpimpl

import (
    "context"
)

func HandleX_farnsworth_syntek(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        name = "Farnsworth Syntek"
    }
    return ok("Hello from " + name + " MCP server")
}