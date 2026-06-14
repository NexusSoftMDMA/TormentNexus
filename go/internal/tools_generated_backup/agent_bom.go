package tools

import (
	"context"
	"fmt"
)

func HandleGetBom(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok(fmt.Sprintf("BOM for %s generated successfully", name))
}

func HandleListBoms(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	return success(fmt.Sprintf("Listed %d BOMs", limit))
}