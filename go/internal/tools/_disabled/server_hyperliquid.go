package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	coin, _ :=getString(args, "coin")
	if coin == "" {
		return err("coin is required")
}

	url := fmt.Sprintf("https://api.hyperliquid.xyz/info?type=allMids")
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch price: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	price, found := data[coin].(string)
	if !found {
		return err("coin not found")
}

	return ok(fmt.Sprintf("Price of %s: %s", coin, price))
}