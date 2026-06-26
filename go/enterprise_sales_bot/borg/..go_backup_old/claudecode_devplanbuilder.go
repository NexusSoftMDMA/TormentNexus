package tools

import (
	"context"
)

func HandleCreateDevplan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	description, _ :=getString(args, "description")
	if description == "" {
		return err("description is required")
}

	plan := "Development plan for: " + description
	return ok(plan)
}