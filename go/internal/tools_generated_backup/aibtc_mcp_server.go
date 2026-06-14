package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	url := fmt.Sprintf("https://mempool.space/api/address/%s", address)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch balance: " + e.Error())
}

	defer resp.Body.Close()
	var data struct {
		ChainStats struct {
			FundedTxoSum int64 `json:"funded_txo_sum"`
			SpentTxoSum  int64 `json:"spent_txo_sum"`
		} `json:"chain_stats"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	balanceSats := data.ChainStats.FundedTxoSum - data.ChainStats.SpentTxoSum
	balanceBtc := float64(balanceSats) / 1e8
	return ok(fmt.Sprintf("Balance: %.8f BTC", balanceBtc))
}