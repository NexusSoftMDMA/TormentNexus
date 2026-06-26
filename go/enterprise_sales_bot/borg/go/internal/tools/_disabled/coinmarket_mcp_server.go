package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleListLatest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	req, e := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/cryptocurrency/listings/latest", nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("X-CMC_PRO_API_KEY", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	data, _ := json.Marshal(result)
	return success(string(data))
}

func HandleQuote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	symbol, _ :=getString(args, "symbol")
	if apiKey == "" || symbol == "" {
		return err("api_key and symbol are required")
}

	req, e := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?symbol="+symbol, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("X-CMC_PRO_API_KEY", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	data, _ := json.Marshal(result)
	return success(string(data))
}