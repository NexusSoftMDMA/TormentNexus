package mcpimpl

import (
	"context"
	"fmt"
)

func HandleGetDeity(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	deities := map[string]string{
		"Zeus":  "King of the gods",
		"Odin":  "All-father",
		"Shiva": "Destroyer",
	}
	info, found := deities[name]
	if !found {
		return err("deity not found")
}

	return ok(info)
}

func HandleListDeities(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	list := []string{"Zeus", "Odin", "Shiva"}
	return ok(fmt.Sprintf("Deities: %v", list))
}