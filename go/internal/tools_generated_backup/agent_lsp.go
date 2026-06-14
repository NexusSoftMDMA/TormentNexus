package tools

import (
	"context"
)

func HandleAnalyze(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code is required")
}

	return success("analysis complete for: " + code)
}

func HandleListTools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("available tools: analyze")
}