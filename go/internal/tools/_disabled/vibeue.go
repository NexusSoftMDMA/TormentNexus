package tools

import (
	"context"
	"fmt"
)

func HandleGenerateCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	desc, _ :=getString(args, "description")
	if desc == "" {
		return err("description is required")
}

	code := fmt.Sprintf("// Vibeue generated code\n// Based on: %s\nvoid VibeueFunction() {\n    // TODO: implement\n}", desc)
	return success(code)
}

func HandleAnalyzeCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code is required")
}

	analysis := fmt.Sprintf("Analyzed %d characters of Unreal Engine code.", len(code))
	return ok(analysis)
}