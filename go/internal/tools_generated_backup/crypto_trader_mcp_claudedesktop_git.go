package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func HandleGetPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	ids := strings.ToLower(symbol)
	apiURL := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd", url.QueryEscape(ids))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	var result map[string]map[string]float64
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	priceData, found := result[ids]
	if !found {
		return err("symbol not found")
}

	price, found := priceData["usd"]
	if !found {
		return err("price not available")
}

	return ok(fmt.Sprintf("Current price of %s is $%.2f USD", symbol, price))
}

func HandleExecuteTrade(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	side, _ :=getString(args, "side")
	quantityStr, _ :=getString(args, "quantity")
	if symbol == "" || side == "" || quantityStr == "" {
		return err("symbol, side, and quantity are required")
}

	quantity, e := strconv.ParseFloat(quantityStr, 64)
	if e != nil {
		return err("invalid quantity")
}

	message := fmt.Sprintf("Executed %s trade for %s: quantity %.4f", side, symbol, quantity)
	return ok(message)
}