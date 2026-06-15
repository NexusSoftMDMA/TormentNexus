package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetAssetPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	url := fmt.Sprintf("https://min-api.cryptocompare.com/data/price?fsym=%s&tsyms=USD", symbol)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch price: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]float64
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid price data")
}

	price, found := result["USD"]
	if !found {
		return err("asset not found")
}

	return ok(fmt.Sprintf("Price of %s is $%.2f", symbol, price))
}