package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetTickets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	subdomain, _ :=getString(args, "subdomain")
	email, _ :=getString(args, "email")
	token, _ :=getString(args, "token")
	if subdomain == "" || email == "" || token == "" {
		return err("subdomain, email, and token are required")
}

	url := fmt.Sprintf("https://%s.zendesk.com/api/v2/tickets.json", subdomain)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(email+"/token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Found tickets: %v", result))
}