package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetHiveAccount(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("username is required")
}

	url := fmt.Sprintf("https://api.hive.blog/1/accounts/%s", username)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return ok(fmt.Sprintf("Account: %+v", result))
}

func HandleGetHivePrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	currency, _ :=getString(args, "currency")
	if currency == "" {
		currency = "usd"
	}
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=hive&vs_currencies=%s", currency)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]map[string]float64
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	price, found := result["hive"][currency]
	if !found {
		return err("price not found")
}

	return ok(fmt.Sprintf("Hive price in %s: %f", currency, price))
}