package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleQueryHoneycomb(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	dataset, _ :=getString(args, "dataset")
	query, _ :=getString(args, "query")
	if apiKey == "" || dataset == "" {
		return err("api_key and dataset are required")
}

	body := strings.NewReader(query)
	req, e := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://api.honeycomb.io/1/events/%s", dataset), body)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("X-Honeycomb-Team", apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("failed to decode response: %v", e))
}

	return ok(fmt.Sprintf("Query result: %v", result))
}