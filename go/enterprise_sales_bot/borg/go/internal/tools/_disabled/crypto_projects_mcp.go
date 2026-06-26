package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListCryptoProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=%d&page=1", limit)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch crypto projects: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("API returned status " + resp.Status)
}

	var projects []struct {
		Name   string  `json:"name"`
		Symbol string  `json:"symbol"`
		Price  float64 `json:"current_price"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&projects); e != nil {
		return err("failed to parse response: " + e.Error())
}

	result := ""
	for _, p := range projects {
		result += fmt.Sprintf("- %s (%s): $%.2f\n", p.Name, p.Symbol, p.Price)

	if result == "" {
		result = "No crypto projects found."
	}
	return ok(result)
}
}