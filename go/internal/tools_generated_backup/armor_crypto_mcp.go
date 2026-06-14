package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetCryptoPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("missing symbol")
	}
	url := "https://api.binance.com/api/v3/ticker/price?symbol=" + symbol
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	var result struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("decode failed: " + e.Error())
	}
	return ok(fmt.Sprintf("price of %s is %s", result.Symbol, result.Price))
}

func HandleGetCrypto24hrTicker(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("missing symbol")
	}
	url := "https://api.binance.com/api/v3/ticker/24hr?symbol=" + symbol
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	var ticker struct {
		Symbol             string `json:"symbol"`
		LastPrice          string `json:"lastPrice"`
		Volume             string `json:"volume"`
		PriceChangePercent string `json:"priceChangePercent"`
	}
	e = json.NewDecoder(resp.Body).Decode(&ticker)
	if e != nil {
		return err("decode failed: " + e.Error())
	}
	msg := fmt.Sprintf("Symbol: %s, Last Price: %s, Volume: %s, Change: %s%%", ticker.Symbol, ticker.LastPrice, ticker.Volume, ticker.PriceChangePercent)
	return ok(msg)
}