//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleIapticListPurchases(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	userId, _ :=getString(args, "user_id")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.iaptic.com/v2/purchases?user_id="+userId, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	result, found := map[string]interface{}{}, false
	e = json.Unmarshal(body, &result)
	if e != nil {
		return err("invalid JSON")
	return success("purchases listed")
}
	_ = found
	return ok("purchases listed")

func HandleIapticGetReceipt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	purchaseId, _ :=getString(args, "purchase_id")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.iaptic.com/v2/receipts/"+purchaseId, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	result, found := map[string]interface{}{}, false
	e = json.Unmarshal(body, &result)
	if e != nil {
		return err("invalid JSON")
}

	_ = found
	return success("receipt retrieved")
}
}