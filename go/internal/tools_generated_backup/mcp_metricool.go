package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleMetricoolGetProfile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	accountID, _ :=getString(args, "account_id")
	if apiKey == "" || accountID == "" {
		return err("missing api_key or account_id")
}

	url := fmt.Sprintf("https://api.metricool.com/v1/accounts/%s?api_key=%s", accountID, apiKey)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

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
		return err(fmt.Sprintf("API error: %s - %s", resp.Status, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Profile data: %s", string(body)))
}

func HandleMetricoolGetInsights(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	accountID, _ :=getString(args, "account_id")
	startDate, _ :=getString(args, "start_date")
	endDate, _ :=getString(args, "end_date")
	if apiKey == "" || accountID == "" {
		return err("missing api_key or account_id")
}

	url := fmt.Sprintf("https://api.metricool.com/v1/accounts/%s/insights?api_key=%s&start_date=%s&end_date=%s", accountID, apiKey, startDate, endDate)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

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
		return err(fmt.Sprintf("API error: %s - %s", resp.Status, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Insights data: %s", string(body)))
}