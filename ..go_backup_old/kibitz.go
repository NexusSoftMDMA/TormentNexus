package tools

import "context"

func HandleKibitz(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		msg = "Hello from Kibitz!"
	}
	return ok(msg)
}