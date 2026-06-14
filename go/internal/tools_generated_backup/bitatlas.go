package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetBitcoinPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = ctx
	currency, _ :=getString(args, "currency")
	if currency == "" {
		currency = "usd"
	}
	resp, e := http.DefaultClient.Get("https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=" + currency)
	if e != nil {
		return err("failed to fetch price: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]map[string]float64
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode: " + e.Error())
}

	price, found := data["bitcoin"][currency]
	if !found {
		return err("currency not found")
}

	return ok(fmt.Sprintf("Bitcoin price in %s: $%.2f", currency, price))
}