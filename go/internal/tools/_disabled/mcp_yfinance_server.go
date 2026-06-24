package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleGetQuote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	u := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s", url.PathEscape(symbol))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to fetch quote: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Chart struct {
			Result []struct {
				Meta struct {
					RegularMarketPrice float64 `json:"regularMarketPrice"`
					Currency          string
				} `json:"meta"`
			} `json:"result"`
		} `json:"chart"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	if len(result.Chart.Result) == 0 {
		return err("no data for symbol")
}

	meta := result.Chart.Result[0].Meta
	return ok(fmt.Sprintf("%s: %.2f %s", symbol, meta.RegularMarketPrice, meta.Currency))
}

func HandleGetInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	u := fmt.Sprintf("https://query2.finance.yahoo.com/v10/finance/quoteSummary/%s?modules=assetProfile", url.PathEscape(symbol))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to fetch info: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		QuoteSummary struct {
			Result []struct {
				AssetProfile struct {
					LongBusinessSummary string `json:"longBusinessSummary"`
					Sector             string
					Industry           string
				} `json:"assetProfile"`
			} `json:"result"`
		} `json:"quoteSummary"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	if len(result.QuoteSummary.Result) == 0 {
		return err("no data for symbol")
}

	profile := result.QuoteSummary.Result[0].AssetProfile
	return ok(fmt.Sprintf("Sector: %s, Industry: %s\nSummary: %s", profile.Sector, profile.Industry, profile.LongBusinessSummary))
}