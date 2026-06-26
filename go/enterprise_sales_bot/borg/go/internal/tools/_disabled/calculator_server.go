package tools

import (
	"context"
	"fmt"
)

func HandleCalculate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	operation, _ :=getString(args, "operation")
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")

	var result int
	switch operation {
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
		return err("unknown operation " + operation)
}

	return ok(fmt.Sprintf("%d", result))
}