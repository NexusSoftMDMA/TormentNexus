package tools

import (
	"context"
	"encoding/json"
)

func HandleListAccounts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	accounts := []map[string]interface{}{
		{"accountId": "123", "name": "Account A"},
		{"accountId": "456", "name": "Account B"},
	}
	data, e := json.Marshal(accounts)
	if e != nil {
		return err("failed to marshal accounts")
}

	return ok(string(data))
}

func HandleListContainers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	accountId, _ :=getString(args, "accountId")
	if accountId == "" {
		return err("accountId is required")
}

	containers := []map[string]interface{}{
		{"containerId": "1", "name": "Container 1"},
		{"containerId": "2", "name": "Container 2"},
	}
	data, e := json.Marshal(containers)
	if e != nil {
		return err("failed to marshal containers")
}

	return ok(string(data))
}