package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetWhitepaper(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	coin, _ :=getString(args, "coin")
	if coin == "" {
		return err("Missing 'coin' parameter")
}

	url := fmt.Sprintf("https://api.whitepapers.io/v1/whitepaper/%s", coin)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Content string `json:"content"`
	}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("Decode failed: " + e.Error())
}

	return success(result.Content)
}

func HandleListWhitepapers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.whitepapers.io/v1/list")
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	var tickers []string
	if e = json.NewDecoder(resp.Body).Decode(&tickers); e != nil {
		return err("Decode failed: " + e.Error())
}

	return success(fmt.Sprintf("Available whitepapers: %v", tickers))
}