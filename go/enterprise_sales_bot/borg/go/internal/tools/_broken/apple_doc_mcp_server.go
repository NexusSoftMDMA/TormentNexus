package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchAppleDocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	apiURL := "https://developer.apple.com/search/search?q=" + url.QueryEscape(query) + "&type=documentation&limit=10&format=json"
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result struct {
		Results []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Description string `json:"description"`
		} `json:"results"`
	}
	if e = json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	if len(result.Results) == 0 {
		return ok("No documentation found for '" + query + "'")
}

	output := fmt.Sprintf("Found %d results for '%s':\n", len(result.Results), query)
	for _, r := range result.Results {
		output += fmt.Sprintf("- %s\n  %s\n  %s\n", r.Title, r.Description, "https://developer.apple.com"+r.URL)

	return ok(output)
}
}