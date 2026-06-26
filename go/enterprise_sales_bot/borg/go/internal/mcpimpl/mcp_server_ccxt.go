package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func handleGetTicker(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	exchange, _ :=getString(args, "exchange")
	symbol, _ :=getString(args, "symbol")
	if exchange == "" || symbol == "" {
		return err("exchange and symbol are required")
}

	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s", symbol)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("fetch failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	price, found := result["price"].(string)
	if !found {
		return err("price not found")
}

	return ok(fmt.Sprintf("Exchange: %s, Symbol: %s, Price: %s", exchange, symbol, price))
}

func handleGetExchanges(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	exchanges := []string{"binance", "coinbase", "kraken", "bitfinex"}
	data, _ := json.Marshal(exchanges)
	return ok(string(data))
}