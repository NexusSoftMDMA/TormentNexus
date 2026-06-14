package tools

import "context"

func HandleGetBearNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	return ok("Note: This is a placeholder note for id: " + id)
}