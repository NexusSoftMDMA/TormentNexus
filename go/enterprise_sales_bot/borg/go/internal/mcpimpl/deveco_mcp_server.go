package mcpimpl

import "context"

func HandleDevecoStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Deveco MCP server is running")
}

func HandleDevecoExecute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	return success("Executed command: " + command)
}