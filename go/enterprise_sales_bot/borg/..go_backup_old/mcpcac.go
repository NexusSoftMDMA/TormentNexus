package tools

import (
	"context"
	"fmt"
	"strings"
)

func HandleGenerateCLI(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	toolName, _ :=getString(args, "tool_name")
	params, _ :=getString(args, "params")
	if toolName == "" {
		return err("tool_name is required")
}

	cmd := toolName
	if params != "" {
		cmd = toolName + " " + strings.ReplaceAll(params, ",", " ")

	return ok(fmt.Sprintf("Generated CLI command: %s", cmd))
}
}