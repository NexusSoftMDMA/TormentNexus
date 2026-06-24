package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetQuote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	resp, e := http.DefaultClient.Get("https://longbridge.global/openapi/quote/v1/stock/quote?symbol=" + symbol)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	price, found := data["price"]
	if !found {
		return err("price not found")
}

	return ok(fmt.Sprintf("Symbol %s price: %v", symbol, price))
}

func HandleSearchStocks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyword, _ :=getString(args, "keyword")
	if keyword == "" {
		return err("keyword is required")
}

	resp, e := http.DefaultClient.Get("https://longbridge.global/openapi/quote/v1/stock/search?keyword=" + keyword)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	stocks, found := data["stocks"]
	if !found {
		return err("stocks not found")
}

	return ok(fmt.Sprintf("Search results for %s: %v", keyword, stocks))
}