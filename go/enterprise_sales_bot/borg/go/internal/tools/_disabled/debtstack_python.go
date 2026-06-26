package tools

import (
	"context"
	"fmt"
)

func HandleGetDebtInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	debtID, _ :=getString(args, "debt_id")
	if debtID == "" {
		return err("debt_id is required")
}

	return ok(fmt.Sprintf("Debt %s: amount $5000, due 2025-06-30", debtID))
}

func HandleCalculatePayment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	amount, _ :=getInt(args, "amount")
	months, _ :=getInt(args, "months")
	if amount <= 0 || months <= 0 {
		return err("amount and months must be positive")
}

	payment := float64(amount) / float64(months)
	return success(fmt.Sprintf("Monthly payment: $%.2f", payment))
}