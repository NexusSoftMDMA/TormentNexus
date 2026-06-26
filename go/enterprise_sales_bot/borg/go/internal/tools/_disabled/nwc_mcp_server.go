package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverURL, _ :=getString(args, "server_url")
	walletID, _ :=getString(args, "wallet_id")
	url := fmt.Sprintf("%s/api/v1/wallets/%s/balance", serverURL, walletID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result struct {
		Balance int `json:"balance"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Balance: %d sat", result.Balance))
}

func HandlePayInvoice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverURL, _ :=getString(args, "server_url")
	walletID, _ :=getString(args, "wallet_id")
	invoice, _ :=getString(args, "invoice")
	url := fmt.Sprintf("%s/api/v1/wallets/%s/pay", serverURL, walletID)
	payload := map[string]string{"invoice": invoice}
	bodyBytes, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", url, io.NopCloser(bytes.NewReader(bodyBytes)))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result struct {
		PaymentHash string `json:"payment_hash"`
		Preimage    string `json:"preimage"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Payment sent. Hash: %s, Preimage: %s", result.PaymentHash, result.Preimage))
}