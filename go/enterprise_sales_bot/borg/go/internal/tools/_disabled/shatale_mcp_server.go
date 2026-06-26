package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCreatePurchase(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	amount, _ :=getInt(args, "amount")
	currency, _ :=getString(args, "currency")
	if currency == "" {
		currency = "USD"
	}
	body, _ := json.Marshal(map[string]interface{}{"amount": amount, "currency": currency})
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.shatale.io/purchases", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error: " + e.Error())
}

	if resp.StatusCode != 201 {
		msg, _ := result["error"].(string)
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, msg))
}

	return ok("purchase created: " + fmt.Sprint(result["id"]))
}

func HandleListVirtualCards(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	status, _ :=getString(args, "status")
	url := "https://api.shatale.io/virtual-cards"
	if status != "" {
		url += "?status=" + status
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var cards []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&cards); e != nil {
		return err("decode error: " + e.Error())
}

	return success(fmt.Sprintf("%d virtual cards retrieved", len(cards)))
}