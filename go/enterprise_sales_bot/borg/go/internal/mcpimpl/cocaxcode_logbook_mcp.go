package mcpimpl

import "context"

func HandleGetNotes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	getString(args, "vault_path")
	return ok("Notes retrieved successfully")
}

func HandleAddNote_cocaxcode_logbook_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	getString(args, "note_content")
	return success("Note added")
}