package tools

import (
	"context"
)

func HandleGetSummary(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	return success("Build summary retrieved for " + path)
}

func HandleGetBundle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	name, _ :=getString(args, "name")
	if path == "" || name == "" {
		return err("path and name are required")
}

	return success("Bundle details for " + name + " at " + path)
}