package mcpimpl

import (
    "context"
)

func HandleAvpHello(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        name = "World"
    }
    return ok("Hello, " + name + " from Avp Sdk!")
}

func HandleAvpEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    message, _ :=getString(args, "message")
    if message == "" {
        return err("message is required")
}

    return ok("Echo: " + message)
}