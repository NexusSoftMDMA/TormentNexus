package tools

import (
	"context"
)

func HandleSync(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "target")
	if target == "" {
		return err("target is required")
}

	return ok("sync started for " + target)
}

func HandleList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resource, _ :=getString(args, "resource")
	if resource == "" {
		return err("resource is required")
}

	return success("listing " + resource)
}