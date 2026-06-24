package tools

import (
	"context"
)

func HandleDaisysMcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return success("Hello, " + name + " from Daisys Mcp!")
}