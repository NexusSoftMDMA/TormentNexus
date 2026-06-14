package tools

import (
	"context"
)

func LintRego(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code is required")
}

	// Simulated lint result
	result := "No issues found"
	if len(code) < 10 {
		result = "Warning: code is too short"
	}
	return ok(result)
}