package mcpimpl

import (
	"context"
)

func HandleGreeting_unifai_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	msg := "Hello, " + name + "!"
	return ok(msg)
}

func HandleReverse_unifai_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	runes := []rune(text)
	n := len(runes)
	for i := 0; i < n/2; i++ {
		runes[i], runes[n-1-i] = runes[n-1-i], runes[i]
	}
	return ok(string(runes))
}