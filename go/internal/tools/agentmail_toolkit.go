//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func HandleSendEmail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	subject, _ :=getString(args, "subject")
	body, _ :=getString(args, "body")
	if to == "" || subject == "" || body == "" {
		return err("missing required fields: to, subject, body")
}

	payload := fmt.Sprintf(`{"to":"%s","subject":"%s","body":"%s"}`, to, subject, body)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.agentmail.com/send", strings.NewReader(payload))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to send email: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("email service returned status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	if status, found := result["status"]; found && status == "sent" {
		return success("email sent successfully")
}

	return ok("email queued")
}

func HandleListEmails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	folder, _ :=getString(args, "folder")
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	params := url.Values{}
	if folder != "" {
		params.Set("folder", folder)

	params.Set("limit", fmt.Sprintf("%d", limit))
	u := "https://api.agentmail.com/list?" + params.Encode()
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to list emails: " + e.Error())
}

	defer resp.Body.Close()
	var emails []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&emails); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(fmt.Sprintf("found %d emails", len(emails)))
}
}
