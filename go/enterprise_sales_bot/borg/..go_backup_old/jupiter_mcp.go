package tools

import "context"

func HandleJupiterPing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("pong from Jupiter MCP")
}

func HandleJupiterEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		message = "no message provided"
	}
	return success("Echo: " + message)
}