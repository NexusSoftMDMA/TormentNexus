package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetTrending_crypto_trending_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.coingecko.com/api/v3/search/trending")
	if e != nil {
		return err("failed to fetch trending: " + e.Error())
}

	defer resp.Body.Close()

	var data struct {
		Coins []struct {
			Item struct {
				Name   string `json:"name"`
				Symbol string `json:"symbol"`
			} `json:"item"`
		} `json:"coins"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response: " + e.Error())
}

	msg := "Trending coins:\n"
	for _, c := range data.Coins {
		msg += c.Item.Name + " (" + c.Item.Symbol + ")\n"
	}
	return ok(msg)
}