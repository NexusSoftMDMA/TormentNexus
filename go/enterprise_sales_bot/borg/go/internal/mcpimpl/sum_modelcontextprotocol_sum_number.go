package mcpimpl

import (
	"context"
	"fmt"
)

func HandleSumNumber(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	sum := a + b
	return success(fmt.Sprintf("The sum of %d and %d is %d", a, b, sum))
}