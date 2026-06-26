package mcpimpl

import (
	"context"
	"strings"
)

func HandleFind_needle_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	needle, _ :=getString(args, "needle")
	haystack, _ :=getString(args, "haystack")
	if strings.Contains(haystack, needle) {
		return ok("needle found in haystack")
}

	return ok("needle not found in haystack")
}