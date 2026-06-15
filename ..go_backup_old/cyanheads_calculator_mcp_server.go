package tools

import (
	"context"
	"fmt"
)

func HandleCalculate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	expr, _ :=getString(args, "expression")
	op, _ :=getString(args, "operation")
	result := fmt.Sprintf("Expression: %s, operation: %s", expr, op)
	return ok(result)
}