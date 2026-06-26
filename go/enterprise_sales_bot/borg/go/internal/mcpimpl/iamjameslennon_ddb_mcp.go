package mcpimpl

import (
    "context"
    "fmt"
)

func HandleGreeting_iamjameslennon_ddb_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        name = "Adventurer"
    }
    return ok(fmt.Sprintf("Greetings, %s! Welcome to D&D Beyond MCP.", name))
}

func HandleSearch_iamjameslennon_ddb_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "query")
    return ok(fmt.Sprintf("Searching for: %s", query))
}