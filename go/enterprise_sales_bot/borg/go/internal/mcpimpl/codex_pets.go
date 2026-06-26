package mcpimpl

import (
	"context"
	"encoding/json"
)

func HandleListPets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pets := []string{"Buddy", "Mittens", "Rex", "Luna", "Charlie"}
	limit, _ :=getInt(args, "limit")
	if limit > 0 && limit < len(pets) {
		pets = pets[:limit]
	}
	b, e := json.Marshal(pets)
	if e != nil {
		return err("failed to marshal pets")
	}
	return ok(string(b))
}