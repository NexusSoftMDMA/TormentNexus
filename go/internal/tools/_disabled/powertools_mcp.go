package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

func HandleListTools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tools := []string{"powertool1", "powertool2", "powertool3"}
	data, e := json.Marshal(tools)
	if e != nil {
		return err("failed to marshal tools")
}

	return ok(string(data))
}

func HandleRunTool(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("command is required")
}

	result := fmt.Sprintf("Executed command: %s", command)
	return success(result)
}