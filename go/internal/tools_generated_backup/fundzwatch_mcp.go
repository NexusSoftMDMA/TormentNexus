package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetStockPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ticker, _ :=getString(args, "ticker")
	if ticker == "" {
		return err("ticker is required")
}

	url := fmt.Sprintf("https://api.example.com/stock/%s/price", ticker)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch price: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	price, found := result["price"].(float64)
	if !found {
		return err("price not found in response")
}

	return success(fmt.Sprintf("Current price of %s is $%.2f", ticker, price))
}

func HandleGetFundInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	url := fmt.Sprintf("https://api.example.com/fund/%s/info", symbol)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch fund info: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	name, found := result["name"].(string)
	if !found {
		return err("name not found in response")
}

	return success(fmt.Sprintf("Fund: %s (%s)", name, symbol))
}