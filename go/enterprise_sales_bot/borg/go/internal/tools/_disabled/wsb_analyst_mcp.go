package tools

import (
	"context"
	"fmt"
)

func HandleWsbAnalyst(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	msg := fmt.Sprintf("WSB Analyst recommends: %s is trending", symbol)
	return success(msg)
}