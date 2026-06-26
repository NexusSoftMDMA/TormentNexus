package mcpimpl

import (
	"context"
	"fmt"
)

func HandleSpend(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	amount, _ :=getString(args, "amount")
	category, _ :=getString(args, "category")
	if amount == "" {
		return err("amount is required")
}

	return ok(fmt.Sprintf("Spent %s on %s", amount, category))
}

func HandleBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Current balance: $100.00")
}