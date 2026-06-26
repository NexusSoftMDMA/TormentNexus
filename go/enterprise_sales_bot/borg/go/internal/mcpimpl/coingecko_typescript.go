package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetCoinPrice_coingecko_typescript(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	coinID, _ :=getString(args, "coin_id")
	vsCurrency, _ :=getString(args, "vs_currency")
	if coinID == "" || vsCurrency == "" {
		return err("coin_id and vs_currency are required")
}

	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=%s", coinID, vsCurrency)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	var result map[string]map[string]float64
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("json decode failed: %v", e))
}

	prices, found := result[coinID]
	if !found {
		return err("coin not found")
}

	price, found := prices[vsCurrency]
	if !found {
		return err("currency not found")
}

	return ok(fmt.Sprintf("Price of %s in %s is %.2f", coinID, vsCurrency, price))
}