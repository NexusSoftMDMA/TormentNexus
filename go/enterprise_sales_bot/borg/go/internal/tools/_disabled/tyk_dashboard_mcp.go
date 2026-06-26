package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListApis(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL := os.Getenv("TYK_DASHBOARD_URL")
	apiKey := os.Getenv("TYK_DASHBOARD_API_KEY")
	if baseURL == "" || apiKey == "" {
		return err("TYK_DASHBOARD_URL and TYK_DASHBOARD_API_KEY must be set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/apis?p=-1", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read response failed: %v", e))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json parse error: %v", e))
}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}

func HandleGetApi(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiID, _ :=getString(args, "apiId")
	if apiID == "" {
		return err("apiId is required")
}

	baseURL := os.Getenv("TYK_DASHBOARD_URL")
	apiKey := os.Getenv("TYK_DASHBOARD_API_KEY")
	if baseURL == "" || apiKey == "" {
		return err("TYK_DASHBOARD_URL and TYK_DASHBOARD_API_KEY must be set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/apis/"+apiID, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read response failed: %v", e))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json parse error: %v", e))
}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}