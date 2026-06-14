package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetCryptopolitanNews(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://cryptopolitan.com/api/news")
	if e != nil {
		return err("failed to fetch news: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Top news: %v", result["articles"]))
}

func HandleGetCryptopolitanPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	url := fmt.Sprintf("https://cryptopolitan.com/api/price/%s", symbol)
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

	price, found := data["price"].(float64)
	if !found {
		return err("price not found in response")
}

	return ok(fmt.Sprintf("Price of %s: $%.2f", symbol, price))
}