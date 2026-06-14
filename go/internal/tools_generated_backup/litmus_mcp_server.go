package tools

import "context"

func HandleLitmus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "Litmus"
	}
	return ok("Hello from " + name + " MCP Server!")
}