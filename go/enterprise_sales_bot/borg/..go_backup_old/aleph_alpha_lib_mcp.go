package tools

import (
	"context"
	"encoding/json"
)

func HandleListComponents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	components := []string{"Button", "Card", "Modal", "Input", "Select"}
	data, e := json.Marshal(components)
	if e != nil {
		return err("failed to marshal components")
}

	return ok(string(data))
}

func HandleGetComponent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name parameter is required")
}

	detail := map[string]string{"name": name, "description": "A reusable UI component from the library"}
	data, e := json.Marshal(detail)
	if e != nil {
		return err("failed to marshal detail")
}

	return ok(string(data))
}