package mcpimpl

import "context"

func HandleGetBudgets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	budgetName, _ :=getString(args, "budget_name")
	if budgetName == "" {
		return ok(`{"budgets":[{"id":"default","name":"My Budget"}]}`)
}

	return ok(`{"budgets":[{"id":"default","name":"` + budgetName + `"}]}`)
}

func HandleGetTransactions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	count, _ :=getInt(args, "count")
	if count <= 0 {
		count = 10
	}
	return ok(`{"transactions":[{"id":"txn1","amount":50.0},{"id":"txn2","amount":25.0}]}`)
}