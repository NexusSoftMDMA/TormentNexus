package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.billinggateway.com/v1/balance", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	var result struct {
		Balance float64 `json:"balance"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
	}
	return ok(fmt.Sprintf("Current balance: %.2f", result.Balance))
}

func HandleListInvoices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.billinggateway.com/v1/invoices", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	var invoices []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&invoices); e != nil {
		return err("failed to parse response: " + e.Error())
	}
	return success(fmt.Sprintf("Found %d invoices", len(invoices)))
}