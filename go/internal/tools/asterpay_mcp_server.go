//go:build ignore
// +build ignore

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleCreatePayment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	amount, _ :=getInt(args, "amount")
	currency, _ :=getString(args, "currency")
	description, _ :=getString(args, "description")
	body, _ := json.Marshal(map[string]interface{}{
		"amount":      amount,
		"currency":    currency,
		"description": description,
	})
	resp, e := http.DefaultClient.Post("https://api.asterpay.com/payments", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("failed to create payment: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(result)
}

func HandleGetBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userID, _ :=getString(args, "userId")
	resp, e := http.DefaultClient.Get("https://api.asterpay.com/balance?userId=" + userID)
	if e != nil {
		return err("failed to get balance: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(result)
}
