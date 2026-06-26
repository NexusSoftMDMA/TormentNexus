package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleGetQuote_mcp_market_data(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	url := "https://api.example.com/quote?symbol=" + symbol
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch quote: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}