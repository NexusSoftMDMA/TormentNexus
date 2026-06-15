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

func HandleListAutomations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("MAILCHIMP_API_KEY")
	if apiKey == "" {
		return err("MAILCHIMP_API_KEY not set")
}

	parts := strings.Split(apiKey, "-")
	if len(parts) != 2 {
		return err("invalid API key format")
}

	dc := parts[1]
	baseURL := fmt.Sprintf("https://%s.api.mailchimp.com/3.0", dc)
	url := baseURL + "/automations"
	qs := ""
	if status := getString(args, "status"); status != "" {
		qs += "status=" + status + "&"
	}
	if count := getInt(args, "count"); count > 0 {
		qs += fmt.Sprintf("count=%d", count)

	if qs != "" {
		url += "?" + strings.TrimSuffix(qs, "&")

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	req.SetBasicAuth("anystring", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response")
}

	return ok(fmt.Sprintf("Found %d automations", int(result["total_items"].(float64))))
}

}
}

func HandleStartAutomation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("MAILCHIMP_API_KEY")
	if apiKey == "" {
		return err("MAILCHIMP_API_KEY not set")
}

	parts := strings.Split(apiKey, "-")
	if len(parts) != 2 {
		return err("invalid API key format")
}

	dc := parts[1]
	baseURL := fmt.Sprintf("https://%s.api.mailchimp.com/3.0", dc)
	id, _ :=getString(args, "automation_id")
	if id == "" {
		return err("automation_id is required")
}

	url := fmt.Sprintf("%s/automations/%s/actions/start", baseURL, id)
	req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
	if e != nil {
		return err("failed to create request")
}

	req.SetBasicAuth("anystring", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	return ok("Automation started successfully")
}