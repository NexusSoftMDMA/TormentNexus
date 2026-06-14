package tools

import (
	"context"
	"fmt"
)

func HandleCreateNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	content, _ :=getString(args, "content")
	return ok(fmt.Sprintf("Note '%s' created with content '%s'", title, content))
}

func HandleListNotes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filter, _ :=getString(args, "filter")
	if filter != "" {
		return ok(fmt.Sprintf("Notes filtered by: %s", filter))
}

	return ok("All notes: [{\"id\":1,\"title\":\"Sample\"}]")
}