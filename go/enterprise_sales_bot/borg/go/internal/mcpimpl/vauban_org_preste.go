package mcpimpl

import (
	"context"
)

func HandlePresteInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		msg = "Preste is ready"
	}
	return ok(msg)
}// touch 1781132143
