//go:build ignore
// +build ignore

package tools

import (
	"context"
)

func HandleAsk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	question, _ :=getString(args, "question")
	return success("You asked: " + question)
}

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}
