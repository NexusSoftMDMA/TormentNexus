package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleGetQuote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	u := "https://query1.finance.yahoo.com/v7/finance/quote?symbols=" + url.QueryEscape(symbol)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("parse failed: " + e.Error())
}

	qr, found := data["quoteResponse"].(map[string]interface{})
	if !found {
		return err("no quote response")
}

	r, found := qr["result"].([]interface{})
	if !found || len(r) == 0 {
		return err("no result")
}

	q, found := r[0].(map[string]interface{})
	if !found {
		return err("invalid quote")
}

	price, _ := q["regularMarketPrice"].(float64)
	name, _ := q["longName"].(string)
	msg := fmt.Sprintf("Symbol: %s | Price: %.2f | Name: %s", symbol, price, name)
	return ok(msg)
}