package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func HandleSendSms(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	message, _ :=getString(args, "message")
	if to == "" {
		return err("missing 'to' parameter")
	}
	if message == "" {
		return err("missing 'message' parameter")
	}
	payload := map[string]string{"to": to, "body": message}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal payload")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "https://mcp.eztexting.com/sms", strings.NewReader(string(body)))
	if e != nil {
		return err("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return err(fmt.Sprintf("API error: %d", resp.StatusCode))
	}
	return success("SMS sent successfully")
}

func HandleGetContacts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	u := "https://mcp.eztexting.com/contacts?limit=" + url.QueryEscape(fmt.Sprintf("%d", limit))
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("failed to create request")
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return err(fmt.Sprintf("API error: %d", resp.StatusCode))
	}
	return success("Contacts retrieved")
}