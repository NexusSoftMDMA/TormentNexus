package tools

import "context"

func HandleCreateEntity(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	entityType, _ :=getString(args, "type")
	if entityType == "" {
		return err("type is required")
}

	return ok("Created entity " + name + " of type " + entityType)
}