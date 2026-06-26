package mcpimpl

import (
	"context"
	"fmt"
)

func HandleGetSchema_constellationdev_types(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok(fmt.Sprintf("Schema for %s: {type: 'object', properties: {}}", name))
}

func HandleListSchemas_constellationdev_types(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Available schemas: user, product, order")
}