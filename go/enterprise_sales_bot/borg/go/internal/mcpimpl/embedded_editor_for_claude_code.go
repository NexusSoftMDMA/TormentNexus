package mcpimpl

import (
	"context"
)

func HandleCreateDiagram_embedded_editor_for_claude_code(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	if title == "" {
		return err("title is required")
}

	return ok("Created diagram: " + title)
}

func HandleCreateNote_embedded_editor_for_claude_code(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	if content == "" {
		return err("content is required")
}

	return ok("Created note: " + content)
}