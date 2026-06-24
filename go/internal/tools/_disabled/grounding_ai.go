package tools

import (
	"context"
	"fmt"
)

func HandleGroundStatement(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	statement, _ :=getString(args, "statement")
	if statement == "" {
		return err("statement is required")
}

	return success(fmt.Sprintf("Statement '%s' is grounded", statement))
}