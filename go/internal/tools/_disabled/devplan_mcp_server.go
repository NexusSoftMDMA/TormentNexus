package tools

import (
	"context"
)

func HandleCreateDevPlan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	description, _ :=getString(args, "description")
	result := map[string]interface{}{
		"id":          "plan-001",
		"name":        name,
		"description": description,
	}
	return success(result)
}