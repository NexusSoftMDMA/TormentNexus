package tools

import "context"

func HandleGetCalls(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	accountId, _ :=getString(args, "account_id")
	return ok("Fetched calls for account: " + accountId)
}