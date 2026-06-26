package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleClawpayCreateInvoice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	amount, _ :=getInt(args, "amount")
	currency, _ :=getString(args, "currency")
	if currency == "" {
		currency = "USD"
	}
	description, _ :=getString(args, "description")

	body := map[string]interface{}{
		"amount":      amount,
		"currency":    currency,
		"description": description,
	}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to encode request: " + e.Error())
}

	resp, e := http.Post("https://api.clawpay.com/invoices", "application/json", bytes.NewReader(jsonBody))
	if e != nil {
		return err("http request failed: " + e.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	return success("Invoice created successfully")
}