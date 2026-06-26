package mcpimpl

import "context"

func HandleList_continue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("List of items retrieved")
}

func HandleExecute_continue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	return ok("Executed: " + command)
}