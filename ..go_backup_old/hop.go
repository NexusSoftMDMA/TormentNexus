package tools

import (
    "context"
    "fmt"
)

func HandleHop(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        name = "World"
    }
    msg := fmt.Sprintf("Hop to %s!", name)
    return ok(msg)
}