package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	query, _ :=getString(args, "query")
	if apiKey == "" || query == "" {
		return err("missing required arguments")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.nile.io/v1/query", nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response")
}

	return success("query executed")
}