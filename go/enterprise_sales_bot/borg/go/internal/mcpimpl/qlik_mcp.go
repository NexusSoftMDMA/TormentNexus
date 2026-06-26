package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"io"
)

func HandleListApps_qlik_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "https://your-qlik-server/api/v1"
	}
	url := fmt.Sprintf("%s/apps", baseURL)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch apps: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned %d: %s", resp.StatusCode, string(body)))
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	jsonBytes, _ := json.Marshal(data)
	return success(string(jsonBytes))
}

func HandleGetApp_qlik_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "https://your-qlik-server/api/v1"
	}
	appID, _ :=getString(args, "app_id")
	if appID == "" {
		return err("app_id is required")
}

	url := fmt.Sprintf("%s/apps/%s", baseURL, appID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch app: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned %d: %s", resp.StatusCode, string(body)))
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	jsonBytes, _ := json.Marshal(data)
	return success(string(jsonBytes))
}