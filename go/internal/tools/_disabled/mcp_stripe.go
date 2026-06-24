package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListCharges(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getString(args, "limit")
	if limit == "" {
		limit = "10"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.stripe.com/v1/charges?limit="+limit, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("STRIPE_API_KEY"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response failed: " + e.Error())
}

	return ok(string(body))
}

func HandleCreatePaymentIntent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	amount, _ :=getString(args, "amount")
	currency, _ :=getString(args, "currency")
	if amount == "" || currency == "" {
		return err("amount and currency are required")
}

	payload := fmt.Sprintf("amount=%s&currency=%s", amount, currency)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.stripe.com/v1/payment_intents", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("STRIPE_API_KEY"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Body = io.NopCloser(nil)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response failed: " + e.Error())
}

	return ok(string(body))
}