package mcpimpl

import (
	"context"
	"fmt"
)

func HandleSearchCourses(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return success(fmt.Sprintf("Searching for courses with query: %s", query))
}

func HandleGetCourse_edubase_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	return success(fmt.Sprintf("Getting course with id: %s", id))
}