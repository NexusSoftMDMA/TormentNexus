package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleListIndexes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.twelvelabs.io/v1/indexes", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("x-api-key", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	return ok(fmt.Sprintf("Indexes: %v", result))
}

func HandleSearchVideos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	indexID, _ :=getString(args, "index_id")
	query, _ :=getString(args, "query")
	if apiKey == "" || indexID == "" || query == "" {
		return err("api_key, index_id, and query are required")
}

	u := fmt.Sprintf("https://api.twelvelabs.io/v1/search?query=%s&index_id=%s", url.QueryEscape(query), url.QueryEscape(indexID))
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("x-api-key", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	return ok(fmt.Sprintf("Search results: %v", result))
}