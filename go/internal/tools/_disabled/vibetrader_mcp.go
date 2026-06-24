package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetMarketData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	url := fmt.Sprintf("https://api.example.com/market?symbol=%s", symbol)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode error: " + e.Error())
}

	return ok(fmt.Sprintf("Market data for %s: %v", symbol, data))
}

func HandlePlaceOrder(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	qty, _ :=getInt(args, "quantity")
	side, _ :=getString(args, "side")
	if symbol == "" || side == "" || qty <= 0 {
		return err("symbol, quantity (>0), and side are required")
}

	payload := map[string]interface{}{
		"symbol":   symbol,
		"quantity": qty,
		"side":     side,
	}
	body, _ := json.Marshal(payload)
	resp, e := http.DefaultClient.Post("https://api.example.com/order", "application/json", nil) // simplified
	if e != nil {
		return err("order failed: " + e.Error())
}

	defer resp.Body.Close()
	return success("Order placed for " + symbol)
}