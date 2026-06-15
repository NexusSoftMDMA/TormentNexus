//go:build ignore
// +build ignore

package tools

/**
 * @file alpha_vantage.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Alpha Vantage financial data.
 * Replaces `av` (uvx av-mcp) STDIO entry in mcp.json.
 *
 * Uses the Alpha Vantage REST API (https://www.alphavantage.co).
 * Improvements over original:
 *  - No uvx/Python dependency.
 *  - Supports: stock quotes, time series, forex, crypto, economic indicators.
 *  - Context-aware with timeout; requires AV_API_KEY.
 */

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const avBaseURL = "https://www.alphavantage.co/query"

func avAPIKey() string {
	if k := os.Getenv("AV_API_KEY"); k != "" {
		return k
	}
	return os.Getenv("ALPHA_VANTAGE_API_KEY")
}

func avGet(ctx context.Context, params map[string]string) (map[string]interface{}, error) {
	apiKey := avAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("AV_API_KEY (or ALPHA_VANTAGE_API_KEY) environment variable is not set")
	}

	q := url.Values{}
	q.Set("apikey", apiKey)
	for k, v := range params {
		q.Set(k, v)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, e := http.NewRequestWithContext(ctx, "GET", avBaseURL+"?"+q.Encode(), nil)
	if e != nil {
		return nil, e
	}

	resp, e := client.Do(req)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Alpha Vantage API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return nil, fmt.Errorf("failed to parse Alpha Vantage response: %v", e)
	}

	// Check for error messages in response
	if msg, ok := result["Information"].(string); ok {
		return nil, fmt.Errorf("Alpha Vantage API info: %s", msg)
	}
	if note, ok := result["Note"].(string); ok {
		return nil, fmt.Errorf("Alpha Vantage API note (rate limit?): %s", note)
	}

	return result, nil
}

// HandleAVGlobalQuote gets the latest price quote for a stock symbol.
// Tool: av_quote
func HandleAVGlobalQuote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ := getString(args, "symbol", "ticker")
	if symbol == "" {
		return err("symbol parameter is required")
	}

	result, e := avGet(ctx, map[string]string{
		"function": "GLOBAL_QUOTE",
		"symbol":   symbol,
	})
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleAVTimeSeries gets historical daily time series data.
// Tool: av_time_series
func HandleAVTimeSeries(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ := getString(args, "symbol", "ticker")
	if symbol == "" {
		return err("symbol parameter is required")
	}

	interval, _ := getString(args, "interval")
	function := "TIME_SERIES_DAILY"
	if interval != "" {
		// Support intraday
		function = "TIME_SERIES_INTRADAY"
	}

	outputSize, _ := getString(args, "outputsize")
	if outputSize == "" {
		outputSize = "compact"
	}

	params := map[string]string{
		"function":   function,
		"symbol":     symbol,
		"outputsize": outputSize,
	}
	if interval != "" {
		params["interval"] = interval
	}

	result, e := avGet(ctx, params)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleAVForexRate gets the current exchange rate between two currencies.
// Tool: av_forex_rate
func HandleAVForexRate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fromCurrency, _ := getString(args, "from_currency", "from", "base")
	toCurrency, _ := getString(args, "to_currency", "to", "quote")
	if fromCurrency == "" || toCurrency == "" {
		return err("from_currency and to_currency parameters are required")
	}

	result, e := avGet(ctx, map[string]string{
		"function":      "CURRENCY_EXCHANGE_RATE",
		"from_currency": fromCurrency,
		"to_currency":   toCurrency,
	})
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleAVCryptoRate gets the current exchange rate for a cryptocurrency.
// Tool: av_crypto_rate
func HandleAVCryptoRate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fromCurrency, _ := getString(args, "symbol", "from_currency", "from")
	toCurrency, _ := getString(args, "to_currency", "to")
	if fromCurrency == "" {
		return err("symbol (or from_currency) parameter is required")
	}
	if toCurrency == "" {
		toCurrency = "USD"
	}

	result, e := avGet(ctx, map[string]string{
		"function":      "CURRENCY_EXCHANGE_RATE",
		"from_currency": fromCurrency,
		"to_currency":   toCurrency,
	})
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleAVSearch searches for symbols matching a keyword.
// Tool: av_symbol_search
func HandleAVSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keywords, _ := getString(args, "keywords", "query", "q")
	if keywords == "" {
		return err("keywords parameter is required")
	}

	result, e := avGet(ctx, map[string]string{
		"function": "SYMBOL_SEARCH",
		"keywords": keywords,
	})
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleAVEconomicIndicator retrieves economic indicator data.
// Tool: av_economic_indicator
func HandleAVEconomicIndicator(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	indicator, _ := getString(args, "indicator", "function")
	if indicator == "" {
		// Default to real GDP
		indicator = "REAL_GDP"
	}

	interval, _ := getString(args, "interval")
	params := map[string]string{"function": indicator}
	if interval != "" {
		params["interval"] = interval
	}

	result, e := avGet(ctx, params)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}
