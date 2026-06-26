package mcpimpl

import (
	"context"
)

func HandleListCapsules(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success(`[{"id":"1","name":"Alpha"},{"id":"2","name":"Beta"}]`)
}

func HandleGetCapsule_capsule(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	return success(`{"id":"` + id + `","name":"Capsule ` + id + `"}`)
}