package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func HandleGetKlines(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	interval, _ :=getString(args, "interval")
	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 100
	}
	url := fmt.Sprintf("https://fapi.binance.com/fapi/v1/klines?symbol=%s&interval=%s&limit=%d", symbol, interval, limit)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch klines: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var klines [][]interface{}
	if e := json.Unmarshal(body, &klines); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Fetched %d klines for %s", len(klines), symbol))
}

func HandleCalculateSma(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pricesRaw := args["prices"]
	period, _ :=getInt(args, "period")
	if period < 1 {
		return err("period must be >= 1")
}

	prices, found := pricesRaw.([]interface{})
	if !found {
		return err("prices must be an array")
}

	if len(prices) < period {
		return err("not enough prices")
}

	var sum float64
	for i := 0; i < period; i++ {
		val, e := strconv.ParseFloat(fmt.Sprintf("%v", prices[i]), 64)
		if e != nil {
			return err("invalid price value")
}

		sum += val
	}
	sma := sum / float64(period)
	return ok(fmt.Sprintf("SMA = %.2f", sma))
}