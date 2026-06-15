package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	coin, _ :=getString(args, "coin")
	if coin == "" {
		return err("coin parameter is required")
}

	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd", coin)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch price: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var data map[string]map[string]float64
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse price data")
}

	if prices, found := data[coin]; found {
		if usd, found := prices["usd"]; found {
			return success(fmt.Sprintf("%s price is $%.2f", coin, usd))

	}
	return err("price not found for " + coin)
}

}

func HandleGetMarketData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	coin, _ :=getString(args, "coin")
	if coin == "" {
		return err("coin parameter is required")
}

	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s?localization=false", coin)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch market data: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse market data")
}

	if name, found := data["name"]; found {
		symbol, _ :=getString(args, "") // dummy, actual from data
		if sym, found := data["symbol"]; found {
			symbol = sym.(string)

		msg := fmt.Sprintf("%s (%s) market data retrieved", name, symbol)
		return ok(msg) // ok returns (ToolResponse, error) with nil error
	}
	return err("coin not found")
}
}