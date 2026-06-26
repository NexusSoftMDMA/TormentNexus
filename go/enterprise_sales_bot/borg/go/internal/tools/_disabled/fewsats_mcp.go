package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleGetBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.fewsats.com/v1/balance", nil)
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

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse: " + e.Error())
}

	balance, found := data["balance"].(float64)
	if !found {
		return err("balance not in response")
}

	return ok(fmt.Sprintf("Balance: %.2f", balance))
}

func HandleCreatePaymentLink(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	amount := getFloat(args, "amount")
	description, _ :=getString(args, "description")
	payload := fmt.Sprintf(`{"amount":%.2f,"description":"%s"}`, amount, description)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.fewsats.com/v1/payment-links", strings.NewReader(payload))
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

	if resp.StatusCode != 201 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse: " + e.Error())
}

	url, found := data["url"].(string)
	if !found {
		return err("url not in response")
}

	return ok(fmt.Sprintf("Payment link created: %s", url))
}

func getFloat(args map[string]interface{}, key string) float64 {
	v, _ := args[key].(float64)
	return v
}