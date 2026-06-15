package tools

import "context"

func HandleList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("List of items retrieved")
}

func HandleExecute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	return ok("Executed: " + command)
}