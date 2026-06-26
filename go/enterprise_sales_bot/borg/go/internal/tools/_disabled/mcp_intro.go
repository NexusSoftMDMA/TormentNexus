package tools

import "context"

func HandleGetIntro(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	msg := "Hello, " + name + "! Welcome to MCP Intro."
	return ok(msg)
}