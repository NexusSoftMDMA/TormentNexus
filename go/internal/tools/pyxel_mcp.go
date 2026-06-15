//go:build ignore
// +build ignore

package tools

import (
    "context"
)

func HandleRunPyxel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    script, _ :=getString(args, "script")
    if script == "" {
        return err("script is required")
    return success("Running Pyxel script: " + script)
}

func HandleGetPyxelVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return ok(map[string]interface{}{
        "description": "Pyxel is a retro game engine for Python",
    }),
}
}