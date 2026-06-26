package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleOpnsenseStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	apiKey, _ :=getString(args, "api_key")
	apiSecret, _ :=getString(args, "api_secret")
	if host == "" || apiKey == "" || apiSecret == "" {
		return err("host, api_key, and api_secret are required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", host+"/api/system/status", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(apiKey, apiSecret)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}