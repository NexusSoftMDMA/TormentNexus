package tools

import (
    "context"
    "fmt"
)

func HandleGreeting(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        name = "World"
    }
    return ok(fmt.Sprintf("Hello, %s! From Hippycampus.", name))
}