package mcpimpl

import "context"

func HandlePing_python_homey_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("pong from Python Homey MCP")
}

func HandleHello_python_homey_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok("Hello, " + name + "!")
}