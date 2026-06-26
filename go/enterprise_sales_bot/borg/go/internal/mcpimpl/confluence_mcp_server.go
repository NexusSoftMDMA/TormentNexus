package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchConfluence(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	baseURL, _ :=getString(args, "baseUrl")
	if baseURL == "" {
		return err("baseUrl is required")
}

	apiToken, _ :=getString(args, "apiToken")
	if apiToken == "" {
		return err("apiToken is required")
}

	reqURL := baseURL + "/rest/api/content/search?cql=" + url.QueryEscape("text~'"+query+"'")
	req, e := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.SetBasicAuth("", apiToken)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	results, found := result["results"].([]interface{})
	if !found {
		return err("no results field in response")
}

	return ok(fmt.Sprintf("Found %d results", len(results)))
}