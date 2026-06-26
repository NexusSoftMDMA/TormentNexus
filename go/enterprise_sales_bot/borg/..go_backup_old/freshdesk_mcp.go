package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListFreshdeskTickets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	apiKey, _ :=getString(args, "api_key")
	if domain == "" || apiKey == "" {
		return err("domain and api_key are required")
}

	url := fmt.Sprintf("https://%s.freshdesk.com/api/v2/tickets", domain)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.SetBasicAuth(apiKey, "X")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	return ok(string(body))
}

func HandleGetFreshdeskTicket(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	apiKey, _ :=getString(args, "api_key")
	ticketID, _ :=getInt(args, "ticket_id")
	if domain == "" || apiKey == "" || ticketID == 0 {
		return err("domain, api_key, and ticket_id are required")
}

	url := fmt.Sprintf("https://%s.freshdesk.com/api/v2/tickets/%d", domain, ticketID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.SetBasicAuth(apiKey, "X")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	return ok(string(body))
}