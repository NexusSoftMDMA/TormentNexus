package mcpimpl

import "context"

func HandleReunion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	return ok("Reunion: " + title)
}