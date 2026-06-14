package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleGetTicker(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	u := fmt.Sprintf("https://api.binance.com/api/v3/ticker/24hr?symbol=%s", url.QueryEscape(symbol))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", e))
}

	defer resp.Body.Close()
	var data struct {
		LastPrice string `json:"lastPrice"`
	}
	if e = json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(fmt.Sprintf("JSON decode failed: %v", e))
}

	return success(fmt.Sprintf("Ticker for %s: %s", symbol, data.LastPrice))
}

func HandleGetOrderBook(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 10
	}
	u := fmt.Sprintf("https://api.binance.com/api/v3/depth?symbol=%s&limit=%d", url.QueryEscape(symbol), limit)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", e))
}

	defer resp.Body.Close()
	var data struct {
		Bids [][2]string `json:"bids"`
		Asks [][2]string `json:"asks"`
	}
	if e = json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(fmt.Sprintf("JSON decode failed: %v", e))
}

	return success(fmt.Sprintf("Order book for %s (top %d):\nBids: %v\nAsks: %v", symbol, limit, data.Bids, data.Asks))
}