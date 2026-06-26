package mcpimpl

import (
	"context"
)

func HandleX_echocontext_factory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return success("Echo: " + msg)
}