package mcpimpl

import (
    "context"
    "fmt"
    "time"
)

func HandleYCli(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        name = "World"
    }
    msg := fmt.Sprintf("Hello, %s! Time is %s", name, time.Now().Format(time.RFC3339))
    return ok(msg)
}