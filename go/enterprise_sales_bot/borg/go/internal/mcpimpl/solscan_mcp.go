package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetAccount_solscan_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
	}
	url := fmt.Sprintf("https://api.solscan.io/account/%s", address)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("http request failed: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
	}
	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("json parse failed: " + e.Error())
	}
	return success(fmt.Sprintf("Account data: %v", data))
}

func HandleGetToken_solscan_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
	}
	url := fmt.Sprintf("https://api.solscan.io/token/%s", address)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("http request failed: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
	}
	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("json parse failed: " + e.Error())
	}
	return success(fmt.Sprintf("Token data: %v", data))
}