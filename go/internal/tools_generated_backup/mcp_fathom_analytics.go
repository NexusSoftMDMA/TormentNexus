package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleListSites(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.usefathom.com/v1/sites", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")
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
		return err("API error: " + string(body))
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	jsonStr, _ := json.MarshalIndent(data, "", "  ")
	return ok(string(jsonStr))
}