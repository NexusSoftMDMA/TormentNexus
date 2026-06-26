package tools

import (
	"context"
	"fmt"
)

func HandleAgentInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok(fmt.Sprintf("Agent %s is active", name))
}

func HandleAgentExecute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("command is required")
}

	return ok(fmt.Sprintf("Executing command: %s", command))
}