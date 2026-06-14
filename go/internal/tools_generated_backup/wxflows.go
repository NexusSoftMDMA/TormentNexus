package tools

import (
	"context"
)

func HandleWxflowsQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	return success("queried: " + query)
}

func HandleWxflowsRun(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	workflow, _ :=getString(args, "workflow")
	if workflow == "" {
		return err("workflow is required")
}

	return success("workflow run: " + workflow)
}