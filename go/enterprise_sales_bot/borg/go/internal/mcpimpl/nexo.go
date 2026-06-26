package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetAccountBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.nexo.com/v1/balance", nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch balance")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	balance, found := result["balance"].(float64)
	if !found {
		return err("balance not found")
}

	return ok(fmt.Sprintf("Balance: %.2f", balance))
}

func HandleGetExchangeRates(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pair, _ :=getString(args, "pair")
	if pair == "" {
		pair = "BTCUSD"
	}
	url := fmt.Sprintf("https://api.nexo.com/v1/rates?pair=%s", pair)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch rates")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	rate, found := result["rate"].(float64)
	if !found {
		return err("rate not found")
}

	return ok(fmt.Sprintf("%s rate: %.4f", pair, rate))
}