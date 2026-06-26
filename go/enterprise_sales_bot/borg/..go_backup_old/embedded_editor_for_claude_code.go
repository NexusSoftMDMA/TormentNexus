package tools

import (
	"context"
)

func HandleCreateDiagram(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	if title == "" {
		return err("title is required")
}

	return ok("Created diagram: " + title)
}

func HandleCreateNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	if content == "" {
		return err("content is required")
}

	return ok("Created note: " + content)
}