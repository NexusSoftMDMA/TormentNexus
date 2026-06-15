package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd", symbol)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result map[string]map[string]float64
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("unmarshal failed: %v", e))
}

	if price, found := result[symbol]["usd"]; found {
		return ok(fmt.Sprintf("Price of %s: $%.2f", symbol, price))
}

	return err("price not found")
}

func HandleGetMarketInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s", symbol)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var coin struct {
		MarketData struct {
			CurrentPrice map[string]float64 `json:"current_price"`
			MarketCap    map[string]float64 `json:"market_cap"`
		} `json:"market_data"`
	}
	if e := json.Unmarshal(body, &coin); e != nil {
		return err(fmt.Sprintf("unmarshal failed: %v", e))
}

	price := coin.MarketData.CurrentPrice["usd"]
	cap := coin.MarketData.MarketCap["usd"]
	return ok(fmt.Sprintf("%s - Price: $%.2f, Market Cap: $%.0f", symbol, price, cap))
}