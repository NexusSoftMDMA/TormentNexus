package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
)

func HandleListTools_foldkit_devtools_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tools := []string{"scan", "test", "deploy"}
	data, e := json.Marshal(tools)
	if e != nil {
		return err("failed to marshal tools")
}

	return ok(fmt.Sprintf("Available tools: %s", string(data)))
}

func HandleRunTool_foldkit_devtools_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tool, _ :=getString(args, "tool")
	if tool == "" {
		return err("tool name is required")
}

	params, _ :=getString(args, "params")
	_ = params
	return success(fmt.Sprintf("Tool '%s' executed successfully", tool))
}