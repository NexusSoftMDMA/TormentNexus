package tools

import (
	"context"
	"fmt"
)

func HandleCalculate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	op, _ :=getString(args, "operation")
	var result int
	switch op {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			return err("division by zero")
}

		result = a / b
	default:
		return err("unknown operation: " + op)
}

	return ok(fmt.Sprintf("%d", result))
}