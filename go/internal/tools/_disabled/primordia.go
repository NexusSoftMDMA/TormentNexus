package tools

import (
	"context"
)

func HandleListTools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("primordia tools: echo, ping")
}

func HandleCallTool(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tool, _ :=getString(args, "tool")
	if tool == "echo" {
		msg, _ :=getString(args, "message")
		return success(msg)
}

	return err("unknown tool: " + tool)
}