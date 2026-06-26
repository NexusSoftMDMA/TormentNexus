package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCryptoPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s", symbol)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch price: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("%s price: %s", result.Symbol, result.Price))
}