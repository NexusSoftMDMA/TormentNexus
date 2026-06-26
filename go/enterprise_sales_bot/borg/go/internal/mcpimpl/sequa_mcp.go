package mcpimpl

import (
	"context"
	"fmt"
)

func HandleHello_sequa_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	msg := fmt.Sprintf("Hello, %s! Welcome to Sequa Mcp.", name)
	return ok(msg)
}

func HandleEcho_sequa_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	return ok(text)
}