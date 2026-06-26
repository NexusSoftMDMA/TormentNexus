package tools

import (
	"context"
)

func HandleAttest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok("Attested: " + msg)
}