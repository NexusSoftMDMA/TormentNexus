package tools

import (
	"context"
	"strings"
)

func HandleEnforceContextMode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	action, _ :=getString(args, "action")
	if action == "" {
		return err("missing action parameter")
	}
	if strings.Contains(action, "shell") || strings.Contains(action, "exec") {
		return err("inefficient tool blocked: use context-mode MCP tools instead")
	}
	return success("action allowed: context-mode verified")
}

func HandleValidateToolUsage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	toolName, _ :=getString(args, "tool")
	if toolName == "" {
		return err("missing tool name")
	}
	inefficient := []string{"bash", "sh", "exec", "run_command"}
	for _, bad := range inefficient {
		if strings.Contains(strings.ToLower(toolName), bad) {
			return err("blocked: switch to context-mode MCP tools")
		}
	}
	return ok("tool usage validated")
}// touch 1781132136
