package tools

import (
    "context"
)

func HandleGreet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        return err("name is required")
}

    return ok("Hello, " + name + "!")
}