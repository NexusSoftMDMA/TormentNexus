package mcpimpl

import (
	"context"
	"fmt"
)

func HandleMomenStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return success("Hello, " + name + " from Momen MCP")
}

func HandleMomenTasks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "project_id")
	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 10
	}
	msg := fmt.Sprintf("Retrieved %d tasks for project %s", limit, projectID)
	return success(msg)
}