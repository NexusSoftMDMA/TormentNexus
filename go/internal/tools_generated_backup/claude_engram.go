package tools

import (
	"context"
	"fmt"
)

func HandleAddMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	if content == "" {
		return err("content is required")
}

	return ok(fmt.Sprintf("Memory added: %s", content))
}

func HandleGetMemories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	return ok(fmt.Sprintf("Returning %d memories", limit))
}