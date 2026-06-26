package tools

import (
	"context"
	"strings"
)

func HandleParity(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	value, _ :=getInt(args, "value")
	if value%2 == 0 {
		return ok("even")
}

	return ok("odd")
}

func HandleCanonical(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	result := strings.TrimSpace(strings.ToLower(input))
	return ok(result)
}