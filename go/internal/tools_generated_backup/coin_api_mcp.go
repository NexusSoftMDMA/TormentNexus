package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetCoinPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	coin, _ :=getString(args, "coin")
	if coin == "" {
		return err("missing 'coin' argument")
}

	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd", coin)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch price: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]map[string]float64
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	data, found := result[coin]
	if !found {
		return err("coin not found")
}

	price, found := data["usd"]
	if !found {
		return err("price not available")
}

	return ok(fmt.Sprintf("Current price of %s: $%.2f USD", coin, price))
}