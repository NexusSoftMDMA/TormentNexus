//go:build ignore
// +build ignore

package tools

import "context"

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("pong")
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	result := "Echo: " + msg
	return success(result)
}
