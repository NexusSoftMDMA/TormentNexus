//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

func HandleGetBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	network, _ :=getString(args, "network")
	if address == "" || network == "" {
		return err("address and network are required")
}

	apiKey := os.Getenv("ALCHEMY_API_KEY")
	if apiKey == "" {
		return err("ALCHEMY_API_KEY not set")
}

	u := fmt.Sprintf("https://%s.g.alchemy.com/v2/%s", network, apiKey)
	reqBody := map[string]string{"jsonrpc": "2.0", "method": "eth_getBalance", "params": []string{address, "latest"}, "id": 1}
	b, e := json.Marshal(reqBody)
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	resp, e := http.DefaultClient.Post(u, "application/json", nil)
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	if result["error"] != nil {
		return err(fmt.Sprintf("alchemy error: %v", result["error"]))
}

	return ok(fmt.Sprintf("Balance: %s", result["result"]))
}

func HandleGetTokenBalances(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	network, _ :=getString(args, "network")
	if address == "" || network == "" {
		return err("address and network are required")
}

	apiKey := os.Getenv("ALCHEMY_API_KEY")
	if apiKey == "" {
		return err("ALCHEMY_API_KEY not set")
}

	u := fmt.Sprintf("https://%s.g.alchemy.com/v2/%s", network, apiKey)
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "alchemy_getTokenBalances",
		"params":  []interface{}{address, "erc20"},
		"id":      1,
	}
	b, e := json.Marshal(reqBody)
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	resp, e := http.DefaultClient.Post(u, "application/json", bytes.NewReader(b))
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	if result["error"] != nil {
		return err(fmt.Sprintf("alchemy error: %v", result["error"]))
}

	tokens, found := result["result"].(map[string]interface{})
	if !found {
		return err("invalid result format")
}

	return ok(fmt.Sprintf("Token balances: %v", tokens["tokenBalances"]))
}
