package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGetPayment_mercadopago_tool(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	paymentID, _ :=getString(args, "payment_id")
	if paymentID == "" {
		return err("payment_id is required")
}

	token := os.Getenv("MERCADOPAGO_ACCESS_TOKEN")
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.mercadopago.com/v1/payments/"+paymentID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	// validate JSON
	var v interface{}
	if e := json.Unmarshal(body, &v); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return ok(string(body))
}

func HandleSearchPayments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 10
	}
	offset, _ :=getInt(args, "offset")
	token := os.Getenv("MERCADOPAGO_ACCESS_TOKEN")
	url := fmt.Sprintf("https://api.mercadopago.com/v1/payments/search?limit=%d&offset=%d", limit, offset)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}