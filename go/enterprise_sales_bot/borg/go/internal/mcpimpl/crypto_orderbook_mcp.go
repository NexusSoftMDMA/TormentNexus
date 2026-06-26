package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetOrderbook(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	limit, _ :=getInt(args, "limit", 10)
	url := fmt.Sprintf("https://api.binance.com/api/v3/depth?symbol=%s&limit=%d", symbol, limit)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch orderbook")
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("orderbook API returned error")
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse orderbook")
}

	return success(fmt.Sprintf("Orderbook: %v", result))
}

func HandleGetTicker_crypto_orderbook_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s", symbol)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch ticker")
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("ticker API returned error")
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse ticker")
}

	return success(fmt.Sprintf("Ticker: %v", result))
}