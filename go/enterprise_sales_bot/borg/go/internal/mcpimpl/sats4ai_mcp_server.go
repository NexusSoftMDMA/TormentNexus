package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetBitcoinPrice_sats4ai_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	currency, _ :=getString(args, "currency")
	if currency == "" {
		currency = "usd"
	}
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=%s", currency)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch price: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("failed to parse response: %v", e))
}

	btc, found := data["bitcoin"].(map[string]interface{})
	if !found {
		return err("unexpected response format")
}

	price, found := btc[currency].(float64)
	if !found {
		return err(fmt.Sprintf("currency %s not found", currency))
}

	return ok(fmt.Sprintf("Bitcoin price in %s: %.2f", currency, price))
}

func HandleGetRequest_sats4ai_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url argument is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return ok(string(body))
}