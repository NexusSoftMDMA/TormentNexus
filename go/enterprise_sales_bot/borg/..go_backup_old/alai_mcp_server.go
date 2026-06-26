package tools

import (
	"context"
)

func HandleAlai(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return ok("Hello from Alai Mcp Server!")
}

	return ok("Alai Mcp Server says: " + msg)
}