package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleSubmitPayment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	amount, _ :=getInt(args, "amount")
	currency, _ :=getString(args, "currency")
	description, _ :=getString(args, "description")
	if amount <= 0 {
		return err("invalid amount")
}

	payload := map[string]interface{}{"amount": amount, "currency": currency, "description": description}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("marshal failed")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.agentgate.eu/payments", bytes.NewBuffer(body))
	if e != nil {
		return err("request failed")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error")
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("submission failed")
}

	return success("payment submitted")
}

func HandleAuthorizePayment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "payment_id")
	approved, _ :=getBool(args, "approved")
	if id == "" {
		return err("missing payment_id")
}

	payload := map[string]interface{}{"payment_id": id, "approved": approved}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("marshal failed")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.agentgate.eu/authorize", bytes.NewBuffer(body))
	if e != nil {
		return err("request failed")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error")
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("authorization failed")
}

	return success("payment authorized")
}