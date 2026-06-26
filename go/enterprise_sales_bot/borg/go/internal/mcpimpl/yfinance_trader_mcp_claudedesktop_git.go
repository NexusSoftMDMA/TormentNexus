package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetStockPrice_yfinance_trader_mcp_claudedesktop_git(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?range=1d&interval=1d", symbol)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	chart, found := result["chart"].(map[string]interface{})
	if !found {
		return err("invalid response structure")
}

	resultArr, found := chart["result"].([]interface{})
	if !found || len(resultArr) == 0 {
		return err("no results")
}

	firstResult, found := resultArr[0].(map[string]interface{})
	if !found {
		return err("invalid result")
}

	meta, found := firstResult["meta"].(map[string]interface{})
	if !found {
		return err("no meta")
}

	price, found := meta["regularMarketPrice"].(float64)
	if !found {
		return err("price not found")
}

	msg := fmt.Sprintf("Current price of %s: %.2f", symbol, price)
	return ok(msg)
}