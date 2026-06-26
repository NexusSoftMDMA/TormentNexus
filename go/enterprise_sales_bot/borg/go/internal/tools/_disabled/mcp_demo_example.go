package tools

import "context"

func HandleDemo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		msg = "Hello from mcp-demo-example"
	}
	return ok(msg)
}