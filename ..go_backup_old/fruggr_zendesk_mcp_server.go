package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func HandleGetTicket(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	subdomain, _ :=getString(args, "subdomain")
	ticketID, _ :=getInt(args, "ticket_id")
	if subdomain == "" || ticketID == 0 {
		return err("missing subdomain or ticket_id")
}

	token := os.Getenv("ZENDESK_TOKEN")
	if token == "" {
		return err("ZENDESK_TOKEN not set")
}

	url := fmt.Sprintf("https://%s.zendesk.com/api/v2/tickets/%d.json", subdomain, ticketID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	req.SetBasicAuth(token + "/token", "X")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("Zendesk API error: %s", string(body)))
}

	return success(string(body))
}

func HandleSearchTickets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	subdomain, _ :=getString(args, "subdomain")
	query, _ :=getString(args, "query")
	if subdomain == "" || query == "" {
		return err("missing subdomain or query")
}

	token := os.Getenv("ZENDESK_TOKEN")
	if token == "" {
		return err("ZENDESK_TOKEN not set")
}

	url := fmt.Sprintf("https://%s.zendesk.com/api/v2/search.json?query=%s", subdomain, strings.ReplaceAll(query, " ", "+"))
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	req.SetBasicAuth(token+"/token", "X")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("Zendesk API error: %s", string(body)))
}

	return success(string(body))
}