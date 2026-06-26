package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGetTicket_sshadows_zendesk_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ticketID, _ :=getInt(args, "ticket_id")
	if ticketID == 0 {
		return err("ticket_id is required")
}

	subdomain := os.Getenv("ZENDESK_SUBDOMAIN")
	if subdomain == "" {
		return err("ZENDESK_SUBDOMAIN not set")
}

	email := os.Getenv("ZENDESK_EMAIL")
	token := os.Getenv("ZENDESK_API_TOKEN")
	url := fmt.Sprintf("https://%s.zendesk.com/api/v2/tickets/%d.json", subdomain, ticketID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(email+"/token", token)
	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	data, found := result["ticket"]
	if !found {
		return err("ticket not found in response")
}

	return ok(fmt.Sprintf("Ticket: %v", data))
}