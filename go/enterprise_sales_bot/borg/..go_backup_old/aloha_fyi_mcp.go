package tools

import (
	"context"
)

func HandleGetAlohaFyi(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return success("Aloha! Welcome to Aloha Fyi MCP.")
}

	return success("Aloha, " + name + "!")
}