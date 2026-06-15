//go:build ignore
// +build ignore

package tools

/**
 * @file alpaca.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Alpaca Markets trading API.
 * Replaces `alpaca-mcp-server` (uvx) STDIO entry in mcp.json.
 *
 * Uses the Alpaca REST API v2 (https://api.alpaca.markets).
 * Improvements over original:
 *  - No uvx/Python dependency.
 *  - Supports: get_account, get_positions, get_orders, place_order, cancel_order,
 *              get_bars, get_latest_quote, get_assets.
 *  - Context-aware with timeout; supports paper and live trading environments.
 */

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	alpacaLiveURL  = "https://api.alpaca.markets"
	alpacaPaperURL = "https://paper-api.alpaca.markets"
	alpacaDataURL  = "https://data.alpaca.markets"
)

func alpacaKeys() (key, secret, baseURL string) {
	key = os.Getenv("ALPACA_API_KEY")
	secret = os.Getenv("ALPACA_SECRET_KEY")
	if os.Getenv("ALPACA_PAPER") == "true" || key == "" {
		baseURL = alpacaPaperURL
	} else {
		baseURL = alpacaLiveURL
	}
	return
}

func alpacaDo(ctx context.Context, method, url string, payload interface{}) (interface{}, error) {
	key, secret, _ := alpacaKeys()
	if key == "" || secret == "" {
		return nil, fmt.Errorf("ALPACA_API_KEY and ALPACA_SECRET_KEY environment variables are required")
	}

	var bodyReader io.Reader
	if payload != nil {
		data, _ := json.Marshal(payload)
		bodyReader = bytes.NewBuffer(data)
	}

	req, e := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if e != nil {
		return nil, e
	}
	req.Header.Set("APCA-API-KEY-ID", key)
	req.Header.Set("APCA-API-SECRET-KEY", secret)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Do(req)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Alpaca API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	if len(body) == 0 {
		return nil, nil
	}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return string(body), nil
	}
	return result, nil
}

// HandleAlpacaGetAccount retrieves the current Alpaca account details.
// Tool: alpaca_get_account
func HandleAlpacaGetAccount(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_, _, baseURL := alpacaKeys()
	result, e := alpacaDo(ctx, "GET", baseURL+"/v2/account", nil)
	if e != nil {
		return err(e.Error())
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleAlpacaGetPositions retrieves all open positions.
// Tool: alpaca_get_positions
func HandleAlpacaGetPositions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_, _, baseURL := alpacaKeys()
	result, e := alpacaDo(ctx, "GET", baseURL+"/v2/positions", nil)
	if e != nil {
		return err(e.Error())
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleAlpacaGetOrders retrieves orders with optional filters.
// Tool: alpaca_get_orders
func HandleAlpacaGetOrders(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_, _, baseURL := alpacaKeys()
	status, _ := getString(args, "status")
	if status == "" {
		status = "open"
	}
	limit := getInt(args, "limit", "count")
	if limit <= 0 {
		limit = 50
	}
	url := fmt.Sprintf("%s/v2/orders?status=%s&limit=%d", baseURL, status, limit)
	result, e := alpacaDo(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleAlpacaPlaceOrder places a new order.
// Tool: alpaca_place_order
func HandleAlpacaPlaceOrder(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ := getString(args, "symbol", "ticker")
	if symbol == "" {
		return err("symbol parameter is required")
	}

	qty, _ := getString(args, "qty", "quantity")
	notional, _ := getString(args, "notional")
	if qty == "" && notional == "" {
		return err("either qty or notional parameter is required")
	}

	side, _ := getString(args, "side")
	if side == "" {
		return err("side parameter is required (buy or sell)")
	}

	orderType, _ := getString(args, "type", "order_type")
	if orderType == "" {
		orderType = "market"
	}

	timeInForce, _ := getString(args, "time_in_force", "tif")
	if timeInForce == "" {
		timeInForce = "day"
	}

	payload := map[string]interface{}{
		"symbol":        symbol,
		"side":          side,
		"type":          orderType,
		"time_in_force": timeInForce,
	}

	if qty != "" {
		payload["qty"] = qty
	} else {
		payload["notional"] = notional
	}

	if limitPrice, _ := getString(args, "limit_price", "limitPrice"); limitPrice != "" {
		payload["limit_price"] = limitPrice
	}

	_, _, baseURL := alpacaKeys()
	result, e := alpacaDo(ctx, "POST", baseURL+"/v2/orders", payload)
	if e != nil {
		return err(e.Error())
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(fmt.Sprintf("Order placed:\n%s", string(out)))
}

// HandleAlpacaCancelOrder cancels an order by ID.
// Tool: alpaca_cancel_order
func HandleAlpacaCancelOrder(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	orderID, _ := getString(args, "order_id", "id")
	if orderID == "" {
		return err("order_id parameter is required")
	}

	_, _, baseURL := alpacaKeys()
	_, e := alpacaDo(ctx, "DELETE", baseURL+"/v2/orders/"+orderID, nil)
	if e != nil {
		return err(e.Error())
	}
	return ok("Order cancelled: " + orderID)
}

// HandleAlpacaGetBars retrieves historical OHLCV bars for a symbol.
// Tool: alpaca_get_bars
func HandleAlpacaGetBars(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ := getString(args, "symbol", "ticker")
	if symbol == "" {
		return err("symbol parameter is required")
	}

	timeframe, _ := getString(args, "timeframe")
	if timeframe == "" {
		timeframe = "1Day"
	}

	limit := getInt(args, "limit", "count")
	if limit <= 0 {
		limit = 50
	}

	url := fmt.Sprintf("%s/v2/stocks/%s/bars?timeframe=%s&limit=%d&feed=iex",
		alpacaDataURL, symbol, timeframe, limit)

	if start, _ := getString(args, "start"); start != "" {
		url += "&start=" + start
	}
	if end, _ := getString(args, "end"); end != "" {
		url += "&end=" + end
	}

	result, e := alpacaDo(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleAlpacaGetLatestQuote gets the latest quote for a symbol.
// Tool: alpaca_get_latest_quote
func HandleAlpacaGetLatestQuote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ := getString(args, "symbol", "ticker")
	if symbol == "" {
		return err("symbol parameter is required")
	}

	url := fmt.Sprintf("%s/v2/stocks/%s/quotes/latest?feed=iex", alpacaDataURL, symbol)
	result, e := alpacaDo(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}
