package tools

import "context"

func HandleListBundles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok(`["bundle-1","bundle-2","bundle-3"]`)
}

func HandleGetBundle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok(name + ": sample bundle description")
}