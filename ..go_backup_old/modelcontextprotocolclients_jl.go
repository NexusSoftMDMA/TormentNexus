package tools

import "context"

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return success("Hello, MCP client! No name provided.")
}

	return success("Hello, " + name + "!")
}