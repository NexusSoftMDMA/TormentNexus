package tools

import "context"

func HandleHindsight(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Hindsight MCP server is operational")
}

func HandleHindsightEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	return ok(text)
}