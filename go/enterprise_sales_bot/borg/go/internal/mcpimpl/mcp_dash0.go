package mcpimpl

import "context"

func HandleGetInfo_mcp_dash0(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok("Hello from Dash0, " + name + "!")
}