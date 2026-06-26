package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleExchangeRates(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.bnm.gov.my/public/exchange-rate"
	if curr, _ :=getString(args, "currency"); curr != "" {
		url += "?currency=" + curr
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("Failed to fetch exchange rates: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	var result struct {
		Data []struct {
			CurrencyCode string `json:"currency_code"`
			Rate         string `json:"rate"`
		} `json:"data"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("Failed to parse JSON: " + e.Error())
}

	var msg string
	for _, item := range result.Data {
		msg += fmt.Sprintf("%s: %s\n", item.CurrencyCode, item.Rate)

	if msg == "" {
		msg = "No exchange rate data available."
	}
	return ok(msg)
}
}