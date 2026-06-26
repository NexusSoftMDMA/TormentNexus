package mcpimpl

import (
	"context"
	"fmt"
)

func HandlePentagonalNumber(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	n, _ :=getInt(args, "n")
	if n < 1 {
		return err("n must be a positive integer")
}

	result := n * (3*n - 1) / 2
	return ok(fmt.Sprintf("The %dth pentagonal number is %d", n, result))
}

func HandleIsPentagonal(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	num, _ :=getInt(args, "number")
	if num < 1 {
		return err("number must be a positive integer")
}

	discriminant := 1 + 24*num
	sqrt := 0
	for sqrt*sqrt < discriminant {
		sqrt++
	}
	if sqrt*sqrt != discriminant || (1+sqrt)%6 != 0 {
		return ok(fmt.Sprintf("%d is not a pentagonal number", num))
}

	return ok(fmt.Sprintf("%d is a pentagonal number", num))
}