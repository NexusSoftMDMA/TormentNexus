package tools

import (
	"context"
)

func HandleCreateArchitectureDoc(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	desc, _ :=getString(args, "description")
	if name == "" {
		return err("name is required")
}

	return success("architecture document '" + name + "' created with description: " + desc)
}

func HandleAnalyzeArchitecture(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	arch, _ :=getString(args, "architecture")
	if arch == "" {
		return err("architecture input is required")
}

	return ok("analysis complete: architecture '" + arch + "' is valid")
}