package mcpimpl

import "context"

func HandleHashnet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		msg = "Hashnet Mcp Js is ready"
	}
	return success(msg)
}