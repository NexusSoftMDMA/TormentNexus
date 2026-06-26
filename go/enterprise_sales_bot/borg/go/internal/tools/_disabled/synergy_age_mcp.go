package tools

import (
    "context"
)

func HandleSynergyAge(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        name = "World"
    }
    return ok("Hello, " + name + "! This is the Synergy Age MCP server.")
}