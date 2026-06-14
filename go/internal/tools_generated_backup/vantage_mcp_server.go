package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleListCostReports(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("VANTAGE_API_KEY")
	if apiKey == "" {
		return err("VANTAGE_API_KEY not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.vantage.sh/v2/cost_reports", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

func HandleGetCostReport(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("token is required")
}

	apiKey := os.Getenv("VANTAGE_API_KEY")
	if apiKey == "" {
		return err("VANTAGE_API_KEY not set")
}

	url := fmt.Sprintf("https://api.vantage.sh/v2/cost_reports/%s", token)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}