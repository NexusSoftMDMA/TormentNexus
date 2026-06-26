package mcpimpl

import (
    "context"
    "fmt"
)

func HandleHello_biel_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        name = "World"
    }
    return ok(fmt.Sprintf("Hello, %s!", name))
}

func HandleAdd_biel_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    a, _ :=getInt(args, "a")
    b, _ :=getInt(args, "b")
    return ok(fmt.Sprintf("%d", a+b))
}