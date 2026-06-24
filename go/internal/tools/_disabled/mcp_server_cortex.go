package tools

import (
	"context"
)

func HandleGetData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok("data for " + name)
}

func HandleAnalyze(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	if input == "" {
		return err("input is required")
}

	return success("analysis result for " + input)
}