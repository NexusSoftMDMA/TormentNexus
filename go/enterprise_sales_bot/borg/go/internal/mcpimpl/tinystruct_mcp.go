package mcpimpl

import (
    "context"
)

func HandleHello_tinystruct_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        name = "World"
    }
    return ok("Hello, " + name + "!")
}

func HandleEcho_tinystruct_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    message, _ :=getString(args, "message")
    if message == "" {
        message = "No message provided"
    }
    return success("Echo: " + message)
}