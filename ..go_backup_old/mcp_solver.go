package tools

import (
	"context"
	"math/big"
)

func HandleCalculate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	n, _ :=getInt(args, "number")
	if n < 0 {
		return err("number must be non-negative")
}

	result := new(big.Int).MulRange(1, int64(n))
	return ok(result.String())
}