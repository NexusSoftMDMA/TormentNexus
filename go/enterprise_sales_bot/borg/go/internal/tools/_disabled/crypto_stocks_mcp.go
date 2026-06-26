package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetStockPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	resp, e := http.DefaultClient.Get("https://api.example.com/stock?symbol=" + symbol)
	if e != nil {
		return err("api request failed")
}

	defer resp.Body.Close()
	var result struct {
		Price float64 `json:"price"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response")
}

	return ok(fmt.Sprintf("Price of %s is $%.2f", symbol, result.Price))
}

func HandleGetCryptoPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	resp, e := http.DefaultClient.Get("https://api.example.com/crypto?symbol=" + symbol)
	if e != nil {
		return err("api request failed")
}

	defer resp.Body.Close()
	var result struct {
		Price float64 `json:"price"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response")
}

	return ok(fmt.Sprintf("Price of %s is $%.2f", symbol, result.Price))
}