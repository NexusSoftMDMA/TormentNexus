package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetBalance_powersun_tron_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	url := fmt.Sprintf("https://api.trongrid.io/v1/accounts/%s", address)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Data []struct {
			Balance int64 `json:"balance"`
		} `json:"data"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("parse error: " + e.Error())
}

	if len(result.Data) == 0 {
		return err("account not found")
}

	return ok(fmt.Sprintf("Balance: %d", result.Data[0].Balance))
}

func HandleGetAccountInfo_powersun_tron_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	url := fmt.Sprintf("https://api.trongrid.io/v1/accounts/%s", address)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Data []struct {
			Balance   int64 `json:"balance"`
			Energy    int64 `json:"energy"`
			Bandwidth int64 `json:"bandwidth"`
		} `json:"data"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("parse error: " + e.Error())
}

	if len(result.Data) == 0 {
		return err("account not found")
}

	info := result.Data[0]
	return ok(fmt.Sprintf("Balance: %d, Energy: %d, Bandwidth: %d", info.Balance, info.Energy, info.Bandwidth))
}