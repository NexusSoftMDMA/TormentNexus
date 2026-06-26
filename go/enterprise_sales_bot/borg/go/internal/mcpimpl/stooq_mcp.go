package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleStockQuote_stooq_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	url := fmt.Sprintf("https://stooq.com/q/l/?s=%s&f=sd2t2ohlcv&h&e=json", symbol)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch data: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	symbols, found := result["symbols"].([]interface{})
	if !found || len(symbols) == 0 {
		return err("no data found for symbol")
}

	first, found := symbols[0].(map[string]interface{})
	if !found {
		return err("invalid symbol data")
}

	name, _ := first["name"].(string)
	closePrice, _ := first["close"].(string)
	if closePrice == "" {
		return err("no closing price available")
}

	msg := fmt.Sprintf("Stock %s: close price %s", name, closePrice)
	return ok(msg)
}