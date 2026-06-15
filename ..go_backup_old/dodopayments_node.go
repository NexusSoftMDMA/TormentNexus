package tools

import (
	"context"
)

func HandleCreatePayment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	amount, _ :=getString(args, "amount")
	currency, _ :=getString(args, "currency")
	return ok("Created payment for " + amount + " " + currency)
}

func HandleGetPayment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	return ok("Retrieved payment " + id)
}