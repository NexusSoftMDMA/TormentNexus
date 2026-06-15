package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetTokenBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	token, _ :=getString(args, "tokenAddress")
	if token == "" {
		return err("tokenAddress is required")
}

	apiKey, _ :=getString(args, "apiKey")
	if apiKey == "" {
		return err("apiKey is required")
}

	url := fmt.Sprintf("https://api.bscscan.com/api?module=account&action=tokenbalance&contractaddress=%s&address=%s&tag=latest&apikey=%s", token, address, apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  string `json:"result"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	if result.Status != "1" {
		return err(fmt.Sprintf("API error: %s", result.Message))
}

	return success(fmt.Sprintf("Balance: %s", result.Result))
}

func HandleGetTransactionStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	txhash, _ :=getString(args, "txhash")
	if txhash == "" {
		return err("txhash is required")
}

	apiKey, _ :=getString(args, "apiKey")
	if apiKey == "" {
		return err("apiKey is required")
}

	url := fmt.Sprintf("https://api.bscscan.com/api?module=transaction&action=getstatus&txhash=%s&apikey=%s", txhash, apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  struct {
			IsError string `json:"isError"`
			ErrDesc string `json:"errDescription"`
		} `json:"result"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	if result.Status != "1" {
		return err(fmt.Sprintf("API error: %s", result.Message))
}

	if result.Result.IsError == "0" {
		return success("Transaction successful")
}

	return success(fmt.Sprintf("Transaction failed: %s", result.Result.ErrDesc))
}