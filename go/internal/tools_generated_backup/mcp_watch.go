package tools

import (
	"context"
)

func HandleWatch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	return ok("Now watching: " + path)
}