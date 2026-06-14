package tools

import (
	"context"
	"fmt"
)

func HandleGetInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return success(fmt.Sprintf("Zuul info for %s", name))
}

func HandleSetThreshold(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	threshold, _ :=getInt(args, "threshold")
	if threshold < 0 {
		return err("threshold must be non-negative")
}

	return ok(fmt.Sprintf("Threshold set to %d", threshold))
}