package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleCreateInvoice_lnbits_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiUrl, _ :=getString(args, "api_url")
	apiKey, _ :=getString(args, "api_key")
	amount, _ :=getInt(args, "amount")
	memo, _ :=getString(args, "memo")

	body := fmt.Sprintf(`{"amount":%d,"memo":"%s"}`, amount, memo)
	req, e := http.NewRequestWithContext(ctx, "POST", strings.TrimRight(apiUrl, "/")+"/api/v1/payments", strings.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("X-Api-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return err("API returned status " + resp.Status)
}

	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("decode failed: " + e.Error())
}

	paymentHash, found := result["payment_hash"].(string)
	if !found {
		return err("no payment_hash in response")
}

	payReq, found := result["payment_request"].(string)
	if !found {
		return err("no payment_request in response")
}

	output := fmt.Sprintf("Invoice created: payment_hash=%s, payment_request=%s", paymentHash, payReq)
	return ok(output)
}

func HandleGetWalletBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiUrl, _ :=getString(args, "api_url")
	apiKey, _ :=getString(args, "api_key")

	req, e := http.NewRequestWithContext(ctx, "GET", strings.TrimRight(apiUrl, "/")+"/api/v1/wallet", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("X-Api-Key", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return err("API returned status " + resp.Status)
}

	var wallet map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&wallet)
	if e != nil {
		return err("decode failed: " + e.Error())
}

	balance, found := wallet["balance"].(float64)
	if !found {
		return err("no balance in response")
}

	name, _ := wallet["name"].(string)
	output := fmt.Sprintf("Wallet: %s, Balance: %.0f msat", name, balance)
	return ok(output)
}