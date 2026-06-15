package tools

import (
    "context"
)

func HandleCreateWidget(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    kind, _ :=getString(args, "type")
    if name == "" || kind == "" {
        return err("name and type are required")
}

    return success("created widget: " + name + " of type " + kind)
}

func HandleRenderWidget(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    id, _ :=getString(args, "id")
    if id == "" {
        return err("widget id is required")
}

    return ok("<div>Widget " + id + " rendered</div>")
}