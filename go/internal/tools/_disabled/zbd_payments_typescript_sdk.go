package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleZbdCreateCharge(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	if apiKey == "" {
		return err("missing apiKey")
}

	amount, _ :=getInt(args, "amount")
	if amount <= 0 {
		return err("amount must be positive")
}

	description, _ :=getString(args, "description")

	payload := map[string]interface{}{
		"amount":      amount,
		"description": description,
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("marshal payload: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.zbd.dev/v1/charges", strings.NewReader(string(body)))
	if e != nil {
		return err(fmt.Sprintf("new request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode response: %v", e))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error: %s", result["message"]))
}

	return success(fmt.Sprintf("Charge created: %v", result["data"]))
}

func HandleZbdGetBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	if apiKey == "" {
		return err("missing apiKey")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.zbd.dev/v1/wallet/balance", nil)
	if e != nil {
		return err(fmt.Sprintf("new request: %v", e))
}

	req.Header.Set("apikey", apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode response: %v", e))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error: %s", result["message"]))
}

	balance, found := result["data"].(map[string]interface{})["balance"]
	if !found {
		return err("balance not found in response")
}

	return success(fmt.Sprintf("Balance: %v", balance))
}