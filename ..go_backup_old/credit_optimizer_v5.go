package tools

import "context"

func HandleCreditOptimizer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	income, _ :=getInt(args, "income")
	debt, _ :=getInt(args, "debt")
	score, _ :=getInt(args, "credit_score")
	if income <= 0 || debt < 0 {
		return err("income must be positive, debt must be non-negative")
}

	ratio := float64(debt) / float64(income) * 100
	msg := ""
	if ratio > 43 {
		msg = "High debt-to-income ratio. Consider reducing debt."
	} else if score < 650 {
		msg = "Credit score could be improved. Try paying bills on time."
	} else {
		msg = "Credit profile looks good."
	}
	return success(msg)
}