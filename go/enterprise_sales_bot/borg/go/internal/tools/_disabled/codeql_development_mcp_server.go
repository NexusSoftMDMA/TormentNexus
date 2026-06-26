package tools

import "context"

func HandleGetHelp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("CodeQL Development MCP Server. Available tools: Help, CreateQuery.")
}

func HandleCreateQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	language, _ :=getString(args, "language")
	return success("Created query: " + name + " for language: " + language)
}