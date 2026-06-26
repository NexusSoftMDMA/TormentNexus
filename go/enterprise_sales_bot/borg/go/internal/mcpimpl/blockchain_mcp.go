package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetBalance_blockchain_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	url := fmt.Sprintf("https://api.blockcypher.com/v1/btc/main/addrs/%s/balance", address)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch balance: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	balance, found := data["balance"].(float64)
	if !found {
		return err("balance not found in response")
}

	return ok("Balance: " + fmt.Sprintf("%.0f", balance))
}

func HandleGetTransaction_blockchain_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	txHash, _ :=getString(args, "tx_hash")
	if txHash == "" {
		return err("tx_hash is required")
}

	url := fmt.Sprintf("https://api.blockcypher.com/v1/btc/main/txs/%s", txHash)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch transaction: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse transaction: " + e.Error())
}

	hash, found := data["hash"].(string)
	if !found {
		return err("hash not found in response")
}

	return ok("Transaction hash: " + hash)
}