package tools

import (
	"context"
	"fmt"
)

func HandleGetRenoun(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok("Renoun: " + name)
}

func HandleListRenoun(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	count, _ :=getInt(args, "count")
	if count > 10 {
		count = 10
	}
	return ok(fmt.Sprintf("Listing %d items", count))
}