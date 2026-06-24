package tools

import (
	"context"
)

func HandleCodexMcpTool(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	return success("Codex Mcp Tool received: " + input)
}