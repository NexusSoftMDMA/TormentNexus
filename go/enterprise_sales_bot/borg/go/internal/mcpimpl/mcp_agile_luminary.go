package mcpimpl

import "context"

func HandleGreet_mcp_agile_luminary(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "Agile Team"
	}
	return success("Hello, " + name + "! Welcome to Agile Luminary.")
}

func HandleInfo_mcp_agile_luminary(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Mcp Agile Luminary: Empowering agile teams with AI insights.")
}