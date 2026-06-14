package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleListSites(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiURL, _ :=getString(args, "api_url")
	apiKey, _ :=getString(args, "api_key")
	if apiURL == "" || apiKey == "" {
		return err("api_url and api_key are required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", apiURL+"/api/v1/websites", nil)
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
		return err("failed to read body: " + e.Error())
}

	return success(string(body))
}