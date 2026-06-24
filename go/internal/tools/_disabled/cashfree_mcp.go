package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleGetBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	url := fmt.Sprintf("https://api.cashfree.com/balance?apiKey=%s", apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed")
	}
	return ok(string(body))
}

func HandleCreatePayout(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	amount, _ :=getString(args, "amount")
	beneficiary, _ :=getString(args, "beneficiary")
	payload := fmt.Sprintf(`{"amount":"%s","beneficiary":"%s"}`, amount, beneficiary)
	resp, e := http.DefaultClient.Post("https://api.cashfree.com/payout", "application/json", strings.NewReader(payload))
	if e != nil {
		return err("payout failed: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed")
	}
	return ok(string(body))
}