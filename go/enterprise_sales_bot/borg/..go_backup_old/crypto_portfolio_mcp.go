package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type coingeckoResponse map[string]map[string]float64

func HandleGetCryptoPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	coin, _ :=getString(args, "coin")
	if coin == "" {
		coin = "bitcoin"
	}
	currency, _ :=getString(args, "currency")
	if currency == "" {
		currency = "usd"
	}
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=%s", coin, currency)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch price: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("API returned status " + resp.Status)
}

	var data coingeckoResponse
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response: " + e.Error())
}

	priceMap, found := data[coin]
	if !found {
		return err("coin not found")
}

	price, found := priceMap[currency]
	if !found {
		return err("currency not found")
}

	return success(fmt.Sprintf("Current %s price: %.2f %s", coin, price, currency))
}