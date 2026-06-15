package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleTrackPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ticker, _ :=getString(args, "ticker")
	if ticker == "" {
		return err("ticker is required")
}

	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd", ticker)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch price: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("API returned status " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response")
}

	if coin, found := data[ticker]; found {
		if prices, found := coin.(map[string]interface{}); found {
			if price, found := prices["usd"]; found {
				return success(fmt.Sprintf("Current price of %s: $%v", ticker, price))

		}
	}
	return err("price not found for " + ticker)
}

}

func HandleGetPriceHistory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ticker, _ :=getString(args, "ticker")
	if ticker == "" {
		return err("ticker is required")
}

	days, _ :=getInt(args, "days")
	if days <= 0 || days > 365 {
		days = 30
	}
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s/market_chart?vs_currency=usd&days=%d", ticker, days)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch history: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("API returned status " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response")
}

	if prices, found := data["prices"]; found {
		return success(fmt.Sprintf("Price history for %s over %d days: %v", ticker, days, prices))
}

	return err("price history not available for " + ticker)
}