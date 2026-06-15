package tools

import "context"

func HandleAlkemi(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	return ok("Alkemi MCP: " + input)
}