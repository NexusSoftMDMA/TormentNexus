package mcpimpl

import (
    "context"
)

func HandleListBugs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return ok("List of bugs: [\"Bug1\", \"Bug2\"]")
}

func HandleGetBug_bugsy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    id, _ :=getString(args, "id")
    if id == "" {
        return err("id is required")
}

    return ok("Bug details for ID: " + id)
}