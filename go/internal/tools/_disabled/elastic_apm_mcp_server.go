package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListApmServices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "baseUrl")
	if baseURL == "" {
		return err("baseUrl is required")
}

	apiKey, _ :=getString(args, "apiKey")
	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/apm/services", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	if apiKey != "" {
		req.Header.Set("Authorization", "ApiKey "+apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("JSON parse error: %v", e))
}

	return ok(fmt.Sprintf("Services: %v", result))
}

}

func HandleListApmTransactions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "baseUrl")
	if baseURL == "" {
		return err("baseUrl is required")
}

	serviceName, _ :=getString(args, "serviceName")
	if serviceName == "" {
		return err("serviceName is required")
}

	apiKey, _ :=getString(args, "apiKey")
	url := fmt.Sprintf("%s/api/apm/services/%s/transactions", baseURL, serviceName)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	if apiKey != "" {
		req.Header.Set("Authorization", "ApiKey "+apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("JSON parse error: %v", e))
}

	return ok(fmt.Sprintf("Transactions: %v", result))
}
}