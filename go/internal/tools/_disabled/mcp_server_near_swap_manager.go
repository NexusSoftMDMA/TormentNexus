package tools

import "context"

func HandleSwap(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tokenIn, _ :=getString(args, "tokenIn")
	tokenOut, _ :=getString(args, "tokenOut")
	amount, _ :=getString(args, "amount")
	if tokenIn == "" || tokenOut == "" || amount == "" {
		return err("Missing required parameters: tokenIn, tokenOut, amount")
}

	return success("Swap initiated from " + tokenIn + " to " + tokenOut + " for " + amount)
}

func HandleCancelSwap(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	swapId, _ :=getString(args, "swapId")
	if swapId == "" {
		return err("swapId is required")
}

	return success("Swap " + swapId + " cancelled")
}