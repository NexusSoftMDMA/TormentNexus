package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetCoinPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	coin, _ :=getString(args, "coin")
	if coin == "" {
		return err("missing 'coin' parameter")
}

	url := "https://api.coingecko.com/api/v3/simple/price?ids=" + coin + "&vs_currencies=usd"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch price: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]map[string]float64
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	priceData, found := result[coin]
	if !found || len(priceData) == 0 {
		return err("coin not found")
}

	price, found := priceData["usd"]
	if !found {
		return err("USD price not available")
}

	return ok("current price of " + coin + " is $" + formatFloat(price))
}

func HandleListCoins(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.coingecko.com/api/v3/coins/list"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch coin list: " + e.Error())
}

	defer resp.Body.Close()
	var list []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&list); e != nil {
		return err("failed to decode list: " + e.Error())
}

	if len(list) > 10 {
		list = list[:10]
	}
	var names []string
	for _, c := range list {
		names = append(names, c.ID+" ("+c.Name+")")

	return ok("available coins (first 10): " + joinStrings(names, ", "))
}
}