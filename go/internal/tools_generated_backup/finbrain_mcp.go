package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetQuote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	token, _ :=getString(args, "api_token")
	if symbol == "" || token == "" {
		return err("symbol and api_token required")
}

	url := fmt.Sprintf("https://api.stockdata.org/v1/data/quote?symbols=%s&api_token=%s", symbol, token)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("API request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read error: %v", e))
}

	return ok(string(body))
}

func HandleGetNews(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	token, _ :=getString(args, "api_token")
	if symbol == "" || token == "" {
		return err("symbol and api_token required")
}

	url := fmt.Sprintf("https://api.stockdata.org/v1/news?symbols=%s&api_token=%s", symbol, token)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("API request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read error: %v", e))
}

	return ok(string(body))
}