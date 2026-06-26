package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetStockQuote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/quote?symbols=%s", symbol)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch quote: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result struct {
		QuoteResponse struct {
			Result []struct {
				Symbol    string  `json:"symbol"`
				LongName  string  `json:"longName"`
				RegularPrice float64 `json:"regularMarketPrice"`
			} `json:"result"`
		} `json:"quoteResponse"`
	}
	if e = json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	if len(result.QuoteResponse.Result) == 0 {
		return err(fmt.Sprintf("no quote found for symbol %s", symbol))
}

	quote := result.QuoteResponse.Result[0]
	msg := fmt.Sprintf("Symbol: %s, Name: %s, Price: %.2f", quote.Symbol, quote.LongName, quote.RegularPrice)
	return ok(msg)
}

func HandleSearchStock(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v1/finance/search?q=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("search failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read: %v", e))
}

	var result struct {
		Quotes []struct {
			Symbol string `json:"symbol"`
			Name   string `json:"shortName"`
		} `json:"quotes"`
	}
	if e = json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse error: %v", e))
}

	if len(result.Quotes) == 0 {
		return err(fmt.Sprintf("no results for %s", query))
}

	var out string
	for _, q := range result.Quotes {
		out += fmt.Sprintf("%s (%s)\n", q.Symbol, q.Name)

	return success(out)
}
}