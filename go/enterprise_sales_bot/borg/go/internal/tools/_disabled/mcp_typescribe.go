package tools

import "context"

func HandleGenerateType(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	return ok("Generated type for: " + input)
}

func HandleDescribeType(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return ok("Describing type: " + name)
}