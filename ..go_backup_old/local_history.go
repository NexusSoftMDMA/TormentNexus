package tools

import (
	"context"
)

func HandleListHistory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	return success("history for " + path)
}

func HandleGetHistory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	file, _ :=getString(args, "file")
	if file == "" {
		return err("file is required")
}

	return ok("found history for " + file)
}