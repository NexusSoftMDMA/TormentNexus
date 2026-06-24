package tools

import (
    "context"
    "fmt"
)

func HandleGreeting(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        name = "Adventurer"
    }
    return ok(fmt.Sprintf("Greetings, %s! Welcome to D&D Beyond MCP.", name))
}

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "query")
    return ok(fmt.Sprintf("Searching for: %s", query))
}