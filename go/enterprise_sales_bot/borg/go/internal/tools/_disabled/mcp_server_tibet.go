package tools

import "context"

func HandleGetInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "Tibet"
	}
	msg := "Hello from " + name + "! MCP Server Tibet is running."
	return ok(msg)
}