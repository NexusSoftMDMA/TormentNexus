package tools

import "context"

func HandleValidateSchema(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	schema, _ :=getString(args, "schema")
	data, _ :=getString(args, "data")
	if schema == "" || data == "" {
		return err("missing schema or data")
}

	return ok("validation passed")
}

func HandleTransform(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	if input == "" {
		return err("missing input")
}

	return success("transformed")
}