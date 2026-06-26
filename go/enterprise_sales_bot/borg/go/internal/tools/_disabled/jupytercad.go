package tools

import (
    "context"
    "fmt"
)

func HandleExecuteCadCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    command, _ :=getString(args, "command")
    if command == "" {
        return err("command is required")
}

    return ok(fmt.Sprintf("Executed CAD command: %s", command))
}