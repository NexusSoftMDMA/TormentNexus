package tools

import (
	"context"
	"strings"
)

func HandleListNotes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	notes := []string{"Welcome", "Getting Started", "Daily Log"}
	return success("Notes: " + strings.Join(notes, ", "))
}

func HandleCreateNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	content, _ :=getString(args, "content")
	if title == "" {
		return err("title is required")
	}
	_ = content
	return ok("Note created: " + title)
}