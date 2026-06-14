package tools

import (
	"context"
	"fmt"
)

func HandleCreateCanvas(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return success(fmt.Sprintf("canvas '%s' created", name))
}

func HandleGetCanvas(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	return ok(fmt.Sprintf(`{"id":"%s","name":"Demo Canvas","nodes":[]}`, id))
}