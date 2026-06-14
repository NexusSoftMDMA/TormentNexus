package tools

import (
	"context"
)

func HandleGetBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	account, _ :=getString(args, "account")
	if account == "" {
		return err("account is required")
}

	return ok("balance for " + account)
}

func HandleListTransactions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	account, _ :=getString(args, "account")
	limit, _ :=getInt(args, "limit")
	if limit < 0 {
		return err("limit must be non-negative")
}

	_ = limit
	if account == "" {
		return err("account is required")
}

	return ok("transactions for " + account)
}