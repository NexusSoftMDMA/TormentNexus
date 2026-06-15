package tools

import (
	"context"
)

func HandleCreateBundle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	desc, _ :=getString(args, "description")
	return success("Created bundle: " + name + " - " + desc)
}

func HandleListBundles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filter, _ :=getString(args, "filter")
	return success("Listed bundles with filter: " + filter)
}