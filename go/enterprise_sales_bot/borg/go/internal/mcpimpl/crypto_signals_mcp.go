package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetCryptoSignal(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	coin, _ :=getString(args, "coin")
	if coin == "" {
		return err("coin is required")
}

	url := "https://api.coingecko.com/api/v3/simple/price?ids=" + coin + "&vs_currencies=usd"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch price: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]map[string]float64
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response")
}

	price, found := data[coin]["usd"]
	if !found {
		return err("coin not found")
}

	var signal string
	switch {
	case price > 50000:
		signal = "buy"
	case price > 30000:
		signal = "hold"
	default:
		signal = "sell"
	}
	return ok(signal)
}