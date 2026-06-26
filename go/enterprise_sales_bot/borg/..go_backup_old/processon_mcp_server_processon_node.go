package tools

import (
	"context"
)

func HandleCreateMindMap(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	md, _ :=getString(args, "markdown")
	if md == "" {
		return err("markdown argument is required")
}

	return success("Mind map created from markdown: " + md)
}