package mcpimpl

import (
    "context"
    "net/http"
)

func HandleAdapt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    app, _ :=getString(args, "app")
    if app == "" {
        return err("missing 'app' argument")
}

    _ = http.DefaultClient
    return ok("Loaded app: " + app)
}