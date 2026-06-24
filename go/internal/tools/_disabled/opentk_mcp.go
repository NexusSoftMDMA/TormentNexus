package tools

import "context"

func HandleOpentkInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok("Hello from Opentk Mcp, " + name + "!")
}