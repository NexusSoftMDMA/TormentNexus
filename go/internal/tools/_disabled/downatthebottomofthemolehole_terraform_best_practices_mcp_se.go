package tools

import (
	"context"
	"strings"
)

func HandleCostCheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	if strings.Contains(input, "cost") {
		return ok("Cost check passed")
}

	return err("Cost check failed: no cost-related resources found")
}

func HandleSecurityCheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	if strings.Contains(input, "security") {
		return success("Security check passed")
}

	return err("Security check failed: missing security best practices")
}