package tools

import (
	"context"
	"encoding/json"
	"strings"
)

func HandleGetMerchantAccount(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok(`{"merchant_account_id":"mck1234","status":"active"}`)
}

func HandleGetCustomer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	customerID, _ :=getString(args, "customer_id")
	if strings.TrimSpace(customerID) == "" {
		return err("customer_id is required")
}

	customer := map[string]interface{}{
		"id":         customerID,
		"first_name": "John",
		"last_name":  "Doe",
		"email":      "john.doe@example.com",
	}
	data, e := json.Marshal(customer)
	if e != nil {
		return err("failed to marshal customer: " + e.Error())
}

	return ok(string(data))
}