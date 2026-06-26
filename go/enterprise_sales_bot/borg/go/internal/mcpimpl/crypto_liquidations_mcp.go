package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetLiquidations_crypto_liquidations_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 10
	}
	url := fmt.Sprintf("https://api.bybit.com/v5/market/liquidations?limit=%d", limit)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch liquidations")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	data, found := result["result"].(map[string]interface{})
	if !found {
		return err("unexpected response format")
}

	jsonBytes, _ := json.Marshal(data)
	return ok(string(jsonBytes))
}

func HandleLiquidationsBySymbol(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	url := fmt.Sprintf("https://api.bybit.com/v5/market/liquidations?symbol=%s", symbol)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch liquidations")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	data, found := result["result"].(map[string]interface{})
	if !found {
		return err("unexpected response format")
}

	jsonBytes, _ := json.Marshal(data)
	return ok(string(jsonBytes))
}