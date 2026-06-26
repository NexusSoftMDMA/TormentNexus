package mcpimpl

import (
	"context"
)

func HandleX_mcp_server_wuye_ai(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	msg := "Hello, " + name + "! Welcome to Wuye Ai."
	return ok(msg)
}

func HandleY_mcp_server_wuye_ai(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "status" {
		return ok("Wuye Ai is operational")
}

	if query == "version" {
		return ok("1.0.0")
}

	return success("Wuye Ai ready")
}