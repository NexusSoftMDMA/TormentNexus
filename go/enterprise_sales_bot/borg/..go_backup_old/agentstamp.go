package tools

import (
	"context"
	"fmt"
)

func HandleListStamps(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit < 1 {
		limit = 10
	}
	return success(fmt.Sprintf("{\"stamps\": [%d sample stamps]}", limit))
}

func HandleCreateStamp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	design, _ :=getString(args, "design")
	if design == "" {
		return err("design is required")
}

	return ok(fmt.Sprintf("stamp '%s' created", design))
}