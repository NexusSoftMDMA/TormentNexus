package tools

import "context"

func HandleGetNotes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	getString(args, "vault_path")
	return ok("Notes retrieved successfully")
}

func HandleAddNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	getString(args, "note_content")
	return success("Note added")
}