package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCreateInvoice_gooodbilling_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	customer, _ :=getString(args, "customer")
	amount, _ :=getInt(args, "amount")
	description, _ :=getString(args, "description")
	body := map[string]interface{}{"customer": customer, "amount": amount, "description": description}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.gooodbilling.com/v1/invoices", bytes.NewReader(jsonBody))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to send request")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return err(fmt.Sprintf("API error: %v", result["error"]))
}

	return ok(fmt.Sprintf("Invoice created with ID: %v", result["id"]))
}

func HandleListInvoices_gooodbilling_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	url := fmt.Sprintf("https://api.gooodbilling.com/v1/invoices?limit=%d", limit)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to send request")
}

	defer resp.Body.Close()
	var invoices []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&invoices); e != nil {
		return err("failed to decode response")
}

	var out string
	for _, inv := range invoices {
		out += fmt.Sprintf("ID: %v, Customer: %v, Amount: %v\n", inv["id"], inv["customer"], inv["amount"])

	return ok(out)
}
}