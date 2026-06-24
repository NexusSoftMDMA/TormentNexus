package tools

import "context"

func HandleWaggle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok("Waggle says: " + msg)
}