package mcpimpl

import (
	"context"
	"fmt"
)

func HandleCalculate_cyanheads_calculator_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	expr, _ :=getString(args, "expression")
	op, _ :=getString(args, "operation")
	result := fmt.Sprintf("Expression: %s, operation: %s", expr, op)
	return ok(result)
}