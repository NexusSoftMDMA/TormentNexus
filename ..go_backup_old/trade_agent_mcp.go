package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleGetCryptoPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		symbol = "BTC"
	}
	url := "https://api.coindesk.com/v1/bpi/currentprice/" + strings.ToUpper(symbol) + ".json"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		BPI map[string]struct {
			Rate float64 `json:"rate_float"`
		} `json:"bpi"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	if rate, found := result.BPI["USD"]; found {
		return ok(fmt.Sprintf("Price of %s: $%.2f", symbol, rate.Rate))
}

	return err("no USD rate found")
}

func HandleExecuteTrade(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	side, _ :=getString(args, "side")
	qty, _ :=getString(args, "quantity")
	if symbol == "" || side == "" || qty == "" {
		return err("missing required parameter")
}

	return ok(fmt.Sprintf("Executed %s: %s %s shares", side, qty, symbol))
}