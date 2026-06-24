package tools

import (
    "context"
    "fmt"
)

func HandlePriorInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        name = "User"
    }
    return ok(fmt.Sprintf("Prior Mcp says hello, %s!", name))
}

func HandlePriorEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    message, _ :=getString(args, "message")
    if message == "" {
        return err("No message provided")
}

    return ok(fmt.Sprintf("Echo: %s", message))
}