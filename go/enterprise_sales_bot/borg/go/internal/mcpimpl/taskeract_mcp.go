package mcpimpl

import "context"

func HandleCreateTask_taskeract_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	if title == "" {
		return err("title is required")
}

	return success("task created: " + title)
}

func HandleGetTask_taskeract_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "id")
	if id == 0 {
		return err("id is required")
}

	return ok("task retrieved")
}