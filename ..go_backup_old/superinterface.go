package tools

import "context"

func HandleGetGreeting(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return success("Hello, " + name)
}