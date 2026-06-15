package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleListInvoices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "api_base")
	token, _ :=getString(args, "api_token")
	if base == "" || token == "" {
		return err("api_base and api_token are required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", base+"/invoices", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return success(result)
}

func HandleGetInvoice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "api_base")
	token, _ :=getString(args, "api_token")
	id, _ :=getString(args, "id")
	if base == "" || token == "" || id == "" {
		return err("api_base, api_token, and id are required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", base+"/invoices/"+id, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return success(result)
}