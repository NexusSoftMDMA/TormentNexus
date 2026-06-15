package tools

import (
	"context"
)

func HandleListNotes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Found 3 notes: Meeting Notes, Shopping List, Ideas")
}

func HandleGetNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	if title == "" {
		return err("title is required")
}

	return ok("Note '" + title + "' content: This is a sample note.")
}