package tools

import (
    "context"
    "fmt"
)

func HandleSayHello(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        name = "World"
    }
    return ok(fmt.Sprintf("Hello, %s!", name))
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    msg, _ :=getString(args, "message")
    return ok(fmt.Sprintf("Echo: %s", msg))
}