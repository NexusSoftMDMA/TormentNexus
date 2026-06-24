package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleCreatePaymentLink(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyID, _ :=getString(args, "key_id")
	keySecret, _ :=getString(args, "key_secret")
	amount, _ :=getInt(args, "amount")
	currency, _ :=getString(args, "currency")
	if keyID == "" || keySecret == "" || amount == 0 {
		return err("missing required args: key_id, key_secret, amount")
}

	if currency == "" {
		currency = "INR"
	}
	body := fmt.Sprintf(`{"amount":%d,"currency":"%s"}`, amount, currency)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.razorpay.com/v1/payment_links", strings.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(keyID, keySecret)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("created payment link: %v", result["short_url"]))
}

func HandleFetchPayment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyID, _ :=getString(args, "key_id")
	keySecret, _ :=getString(args, "key_secret")
	paymentID, _ :=getString(args, "payment_id")
	if keyID == "" || keySecret == "" || paymentID == "" {
		return err("missing required args: key_id, key_secret, payment_id")
}

	url := fmt.Sprintf("https://api.razorpay.com/v1/payments/%s", paymentID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(keyID, keySecret)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	data, _ := json.Marshal(result)
	return success(string(data))
}