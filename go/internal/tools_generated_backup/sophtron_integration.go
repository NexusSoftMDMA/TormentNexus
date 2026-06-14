package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	accountID, _ :=getString(args, "account_id")
	if accountID == "" {
		return err("account_id is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.sophtron.com/v1/accounts/%s/balance", accountID), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("api call failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return success(fmt.Sprintf("Balance: %v", result["balance"]))
}