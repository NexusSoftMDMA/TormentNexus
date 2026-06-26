package mcpimpl

import (
    "context"
)

func HandlePmpt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    prompt, _ :=getString(args, "prompt")
    if prompt == "" {
        return err("prompt is required")
}

    result := "Result: " + prompt
    return ok(result)
}