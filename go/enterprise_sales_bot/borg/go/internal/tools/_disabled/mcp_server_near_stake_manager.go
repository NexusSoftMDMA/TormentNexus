package tools

import "context"

func HandleStake(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	account, _ :=getString(args, "account_id")
	amount, _ :=getInt(args, "amount")
	if account == "" {
		return err("account_id is required")
}

	if amount <= 0 {
		return err("amount must be positive")
}

	return success("stake placed successfully for " + account)
}

func HandleUnstake(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	account, _ :=getString(args, "account_id")
	amount, _ :=getInt(args, "amount")
	if account == "" {
		return err("account_id is required")
}

	if amount <= 0 {
		return err("amount must be positive")
}

	return success("unstake requested for " + account)
}