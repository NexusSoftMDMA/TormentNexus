package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleSendEmail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "baseURL")
	token, _ :=getString(args, "token")
	to, _ :=getString(args, "to")
	subject, _ :=getString(args, "subject")
	body, _ :=getString(args, "body")
	if baseURL == "" || token == "" || to == "" || subject == "" || body == "" {
		return err("Missing required fields: baseURL, token, to, subject, body")
}

	payload := map[string]interface{}{
		"to":      to,
		"subject": subject,
		"body":    body,
	}
	bodyBytes, e := json.Marshal(payload)
	if e != nil {
		return err("Failed to marshal payload: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", baseURL+"/send", strings.NewReader(string(bodyBytes)))
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(respBody)))
}

	return ok("Email sent successfully")
}

func HandleListCampaigns(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "baseURL")
	token, _ :=getString(args, "token")
	if baseURL == "" || token == "" {
		return err("Missing required fields: baseURL, token")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/campaigns", nil)
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(respBody)))
}

	return success(string(respBody))
}