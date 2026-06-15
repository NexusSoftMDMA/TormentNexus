//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
)

func HandleGreeting(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	msg := fmt.Sprintf("Hello, %s! Welcome to Anthropic MCP.", name)
	return ok(msg)
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		msg = "No message provided"
	}
	return ok(msg)
}
