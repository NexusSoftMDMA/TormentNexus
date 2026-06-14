package tools

import "context"

func HandleTailtestCline(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Tailtest Cline MCP server is running.")
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(msg)
}