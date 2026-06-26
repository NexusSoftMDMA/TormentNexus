package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func GetUtxos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	resp, e := http.DefaultClient.Get("https://blockstream.info/api/address/" + address + "/utxo")
	if e != nil {
		return err("failed to fetch UTXOs: " + e.Error())
}

	defer resp.Body.Close()
	var utxos []struct {
		Txid   string `json:"txid"`
		Vout   int    `json:"vout"`
		Value  int64  `json:"value"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&utxos); e != nil {
		return err("failed to decode: " + e.Error())
}

	return ok(fmt.Sprintf("Found %d UTXOs", len(utxos)))
}

func GetBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	resp, e := http.DefaultClient.Get("https://blockstream.info/api/address/" + address + "/utxo")
	if e != nil {
		return err("failed to fetch UTXOs: " + e.Error())
}

	defer resp.Body.Close()
	var utxos []struct {
		Value int64 `json:"value"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&utxos); e != nil {
		return err("failed to decode: " + e.Error())
}

	var total int64
	for _, u := range utxos {
		total += u.Value
	}
	return ok(fmt.Sprintf("Balance: %d satoshi", total))
}