package tools

import (
	"context"
	"fmt"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return success(fmt.Sprintf("Created Xmind mind map for %s", name))
}