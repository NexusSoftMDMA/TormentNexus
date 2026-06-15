package tools

import "context"

func HandleAddNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	content, _ :=getString(args, "content")
	if title == "" || content == "" {
		return err("title and content are required")
}

	return ok("note added: " + title)
}

func HandleListNotes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	return ok("listing up to " + string(rune(limit)) + " notes")
}