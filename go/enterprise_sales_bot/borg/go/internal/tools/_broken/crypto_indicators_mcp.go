package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetCryptoPrice_crypto_indicators_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	currency, _ :=getString(args, "currency")
	if currency == "" {
		currency = "usd"
	}
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=%s", symbol, currency)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var data map[string]map[string]float64
	e = json.Unmarshal(body, &data)
	if e != nil {
		return err(e.Error())
}

	if val, found := data[symbol]; found {
		if price, found2 := val[currency]; found2 {
			return ok(fmt.Sprintf("Price of %s in %s: %.2f", symbol, currency, price))

	}
	return err("price not found")
}
}