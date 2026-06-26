package tools

import (
	"context"
	"fmt"
)

func HandleDeploy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok(fmt.Sprintf("Deployment initiated for %s", name))
}