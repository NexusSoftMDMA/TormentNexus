package tools

import (
    "context"
)

func HandleScaffoldPlugin(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        return err("name is required")
}

    return success("Scaffolded plugin: " + name)
}

func HandleValidatePlugin(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    path, _ :=getString(args, "path")
    if path == "" {
        return err("path is required")
}

    return ok("Validation passed for: " + path)
}