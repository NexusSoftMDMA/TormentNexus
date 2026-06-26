package mcpimpl

import (
	"context"
	"strings"
)

func HandleCleanCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code argument is required")
}

	cleaned := strings.TrimSpace(code)
	return ok(cleaned)
}