package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGetAccount_brightdata_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("BRIGHTDATA_API_KEY")
	if apiKey == "" {
		return err("BRIGHTDATA_API_KEY not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.brightdata.com/account", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read body: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("Account info: %v", result))
}

func HandleGetZones(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("BRIGHTDATA_API_KEY")
	if apiKey == "" {
		return err("BRIGHTDATA_API_KEY not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.brightdata.com/zones", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read body: %v", e))
}

	var zones []interface{}
	if e := json.Unmarshal(body, &zones); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("Zones: %v", zones))
}