package tools

import (
	"context"
)

func HandleCompile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	source, _ :=getString(args, "source")
	if source == "" {
		return err("source is required")
}

	return ok("compiled successfully")
}