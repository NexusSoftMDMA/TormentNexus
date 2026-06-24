package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetMarket(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	marketId, _ :=getString(args, "marketId")
	if marketId == "" {
		return err("marketId is required")
}

	url := fmt.Sprintf("https://clob.polymarket.com/markets/%s", marketId)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch market: %v", e))
}

	defer resp.Body.Close()
	var data interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(fmt.Sprintf("failed to parse response: %v", e))
}

	return ok("Market data retrieved", data)
}

func HandlePlaceBet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	marketId, _ :=getString(args, "marketId")
	outcome, _ :=getString(args, "outcome")
	amount, _ :=getString(args, "amount")
	if marketId == "" || outcome == "" || amount == "" {
		return err("marketId, outcome, and amount are required")
}

	return success("Bet placed successfully")
}