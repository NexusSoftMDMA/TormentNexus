package tools

import (
	"context"
	"fmt"
)

func HandleEval(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	expr, _ :=getString(args, "expression")
	if expr == "" {
		return err("expression argument is required")
}

	result := len(expr)
	return ok(fmt.Sprintf("Expression length: %d", result))
}

func HandleView(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id argument is required")
}

	return ok(fmt.Sprintf("View for ID '%s'", id))
}