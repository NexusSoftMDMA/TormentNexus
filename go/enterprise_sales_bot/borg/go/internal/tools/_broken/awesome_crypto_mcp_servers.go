package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCryptoPrice_awesome_crypto_mcp_servers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		symbol = "BTC"
	}
	currency, _ :=getString(args, "currency")
	if currency == "" {
		currency = "USD"
	}
	url := fmt.Sprintf("https://api.coinbase.com/v2/prices/%s-%s/spot", symbol, currency)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch price")
}

	defer resp.Body.Close()
	var result struct {
		Data struct {
			Amount string `json:"amount"`
		} `json:"data"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	return ok(fmt.Sprintf("Current price of %s in %s is %s", symbol, currency, result.Data.Amount))
}

func HandleGetTopVolume(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	resp, e := http.DefaultClient.Get("https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&order=volume_desc&per_page=" + fmt.Sprint(limit))
	if e != nil {
		return err("failed to fetch volume data")
}

	defer resp.Body.Close()
	var coins []struct {
		Name   string  `json:"name"`
		Symbol string  `json:"symbol"`
		Volume float64 `json:"total_volume"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&coins); e != nil {
		return err("failed to decode response")
}

	msg := fmt.Sprintf("Top %d coins by volume:\n", limit)
	for _, c := range coins {
		msg += fmt.Sprintf("%s (%s): $%.2f\n", c.Name, c.Symbol, c.Volume)

	return ok(msg)
}
}