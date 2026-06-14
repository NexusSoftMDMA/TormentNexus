package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleCreateInvoice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	clientName, _ :=getString(args, "client_name")
	amount, _ :=getString(args, "amount")
	apiKey := os.Getenv("FAKTURKA_API_KEY")
	if apiKey == "" {
		return err("FAKTURKA_API_KEY not set")
}

	body, _ := json.Marshal(map[string]string{"client_name": clientName, "amount": amount})
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.fakturka.pl/v1/invoices", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return ok(fmt.Sprintf("invoice created: %v", result["id"]))
}

func HandleListInvoices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("FAKTURKA_API_KEY")
	if apiKey == "" {
		return err("FAKTURKA_API_KEY not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.fakturka.pl/v1/invoices", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	var invoices []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&invoices)
	return success(invoices)
}