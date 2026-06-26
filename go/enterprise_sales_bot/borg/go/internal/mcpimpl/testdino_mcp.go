package mcpimpl

import (
	"context"
	"fmt"
)

func HandleListDinos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dinos := []string{"T-Rex", "Stegosaurus", "Triceratops"}
	return ok(fmt.Sprintf("Available dinos: %v", dinos))
}

func HandleGetDino(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return success(fmt.Sprintf("Dino: %s, weight: 5000kg, height: 4m", name))
}