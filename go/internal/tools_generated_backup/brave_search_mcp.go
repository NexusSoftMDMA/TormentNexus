package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	apiKey := os.Getenv("BRAVE_API_KEY")
	if apiKey == "" {
		return err("BRAVE_API_KEY not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.search.brave.com/res/v1/web/search?q="+query, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	results, found := result["web"].(map[string]interface{})["results"].([]interface{})
	if !found {
		return success("no results found")
}

	return ok(fmt.Sprintf("Found %d results", len(results)))
}