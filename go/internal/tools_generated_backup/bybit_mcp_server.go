package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	apiSecret, _ :=getString(args, "api_secret")
	if apiKey == "" || apiSecret == "" {
		return err("api_key and api_secret are required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.bybit.com/v5/account/wallet-balance", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("X-BAPI-API-KEY", apiKey)
	req.Header.Set("X-BAPI-SIGN", apiSecret)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return success(fmt.Sprintf("Balance: %v", result))
}

func HandleGetPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	url := fmt.Sprintf("https://api.bybit.com/v5/market/tickers?category=spot&symbol=%s", symbol)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return success(fmt.Sprintf("Price for %s: %v", symbol, result))
}