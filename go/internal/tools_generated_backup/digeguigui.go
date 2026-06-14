package tools

import "context"

func HandleDigeguigui(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return ok("Welcome to Digeguigui MCP server!")
}

	return ok("You said: " + message)
}