package mcpimpl

import "context"

func HandleSqlaugur(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "Sqlaugur"
	}
	return ok("Welcome to " + name + " MCP server.")
}