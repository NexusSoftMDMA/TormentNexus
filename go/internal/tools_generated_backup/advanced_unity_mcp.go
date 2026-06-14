package tools

import (
    "context"
)

func HandleGetUnityVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    version, _ :=getString(args, "version")
    if version == "" {
        version = "2022.3.1f1"
    }
    return ok("Unity version: " + version)
}

func HandleExecuteUnityCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    command, _ :=getString(args, "command")
    if command == "" {
        return err("command argument is required")
}

    return success("Executed command: " + command)
}