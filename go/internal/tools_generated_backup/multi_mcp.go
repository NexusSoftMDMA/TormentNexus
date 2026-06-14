package tools

import (
	"context"
)

func HandleMulti(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "input")
	return success("Processed: " + msg)
}