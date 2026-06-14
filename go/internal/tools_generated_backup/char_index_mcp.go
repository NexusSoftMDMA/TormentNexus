package tools

import (
	"context"
)

func HandleGetCharacter(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("missing name")
}

	return ok("Character: " + name)
}

func HandleListCharacters(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit < 1 {
		limit = 5
	}
	return ok("List of up to " + string(rune(limit)) + " characters")
}