package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetHealthData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	metric, _ :=getString(args, "metric")
	date, _ :=getString(args, "date")
	apiKey, _ :=getString(args, "api_key")
	if metric == "" || apiKey == "" {
		return err("metric and api_key are required")
}

	url := fmt.Sprintf("https://api.vitaltrends.com/v1/health?metric=%s&date=%s", metric, date)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
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
		return err("invalid JSON: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

func HandleGetTrends_vitaltrends_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	metric, _ :=getString(args, "metric")
	startDate, _ :=getString(args, "start_date")
	endDate, _ :=getString(args, "end_date")
	apiKey, _ :=getString(args, "api_key")
	if metric == "" || apiKey == "" || startDate == "" || endDate == "" {
		return err("metric, start_date, end_date, and api_key are required")
}

	url := fmt.Sprintf("https://api.vitaltrends.com/v1/trends?metric=%s&start_date=%s&end_date=%s", metric, startDate, endDate)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
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
		return err("invalid JSON: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}