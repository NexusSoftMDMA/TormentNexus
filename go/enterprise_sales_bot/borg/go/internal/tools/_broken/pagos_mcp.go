package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetPayment_pagos_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing id")
}

	url := fmt.Sprintf("https://api.example.com/pagos/%s", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Payment %s: %v", id, data))
}

func HandleCreatePayment_pagos_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	amount, _ :=getInt(args, "amount")
	if amount <= 0 {
		return err("invalid amount")
}

	body := map[string]interface{}{"amount": amount}
	payload, e := json.Marshal(body)
	if e != nil {
		return err("marshal failed: " + e.Error())
}

	resp, e := http.DefaultClient.Post("https://api.example.com/pagos", "application/json", bytes.NewReader(payload))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	return success("Created payment of " + fmt.Sprint(amount))
}