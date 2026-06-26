package mcpimpl

import "context"

func HandleNcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "msg")
	if msg == "" {
		msg = "Hello from NCP server"
	}
	return success(msg)
}