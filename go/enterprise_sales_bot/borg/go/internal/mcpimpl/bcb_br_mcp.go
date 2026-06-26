package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetExchangeRate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	currency, _ :=getString(args, "currency")
	if currency == "" {
		currency = "USD-BRL"
	}
	url := fmt.Sprintf("https://economia.awesomeapi.com.br/json/last/%s", currency)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch exchange rate")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON")
}

	data, found := result[currency].(map[string]interface{})
	if !found {
		return err("currency data not found")
}

	bid, _ := data["bid"].(string)
	if bid == "" {
		return err("bid value not available")
}

	return ok(fmt.Sprintf("Exchange rate %s: %s", currency, bid))
}

func HandleGetSelicRate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.bcb.gov.br/dados/serie/bcdata.sgs/11/dados/ultimos/1?formato=json"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch Selic rate")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var results []map[string]interface{}
	if e := json.Unmarshal(body, &results); e != nil {
		return err("failed to parse JSON")
}

	if len(results) == 0 {
		return err("no data found")
}

	valor, found := results[0]["valor"].(string)
	if !found {
		return err("value not found in response")
}

	return ok(fmt.Sprintf("Selic rate: %s%%", valor))
}