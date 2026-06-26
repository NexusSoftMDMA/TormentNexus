package tools

import (
	"context"
	"fmt"
)

func HandleGetUniswapPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tokenIn, _ :=getString(args, "tokenIn")
	tokenOut, _ :=getString(args, "tokenOut")
	if tokenIn == "" {
		tokenIn = "ETH"
	}
	if tokenOut == "" {
		tokenOut = "USDC"
	}
	price := 1234.56
	msg := fmt.Sprintf("Uniswap price %s/%s: %f", tokenIn, tokenOut, price)
	return ok(msg)
}