package tools

import (
	"context"
)

func HandleHello(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "Guest"
	}
	return ok("Hello, " + name + "! Welcome to Toreador Mcp Server.")
}

func HandleGoodbye(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "Guest"
	}
	return ok("Goodbye, " + name + "! See you later.")
}