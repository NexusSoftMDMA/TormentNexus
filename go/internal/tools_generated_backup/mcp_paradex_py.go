package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetMarkets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "https://api.paradex.fi/v1"
	}
	resp, e := http.DefaultClient.Get(baseURL + "/markets")
	if e != nil {
		return err("failed to fetch markets: " + e.Error())
}

	defer resp.Body.Close()
	var data interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode markets: " + e.Error())
}

	return ok(data)
}

func HandleGetAccountInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "https://api.paradex.fi/v1"
	}
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	resp, e := http.DefaultClient.Get(baseURL + "/account/" + address)
	if e != nil {
		return err("failed to fetch account: " + e.Error())
}

	defer resp.Body.Close()
	var data interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode account: " + e.Error())
}

	return ok(data)
}