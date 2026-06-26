package mcpimpl

import "context"

func HandleChainlit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	return success("Received: " + message)
}