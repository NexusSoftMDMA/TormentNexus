package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleAddressRisk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	apiKey, _ :=getString(args, "apiKey")
	if address == "" {
		return err("address is required")
}

	u := fmt.Sprintf("https://api.anchain.ai/api/v1/address/risk?address=%s&api_key=%s", url.QueryEscape(address), url.QueryEscape(apiKey))
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("failed to execute request: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("Risk data: %v", result))
}

func HandleTransactionAnalysis(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	txHash, _ :=getString(args, "txHash")
	apiKey, _ :=getString(args, "apiKey")
	if txHash == "" {
		return err("txHash is required")
}

	u := fmt.Sprintf("https://api.anchain.ai/api/v1/transaction/analysis?tx_hash=%s&api_key=%s", url.QueryEscape(txHash), url.QueryEscape(apiKey))
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("failed to execute request: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("Transaction analysis: %v", result))
}