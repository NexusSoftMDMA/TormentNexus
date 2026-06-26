package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type kagiResult struct {
	Data struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Snippet string `json:"snippet"`
	} `json:"data"`
}

type kagiResponse struct {
	Results []kagiResult `json:"results"`
}

func HandleKagiSearch_mcp_servers_kagi(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	u := fmt.Sprintf("https://kagi.com/api/v0/search?q=%s", url.QueryEscape(query))
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	var kResp kagiResponse
	if e := json.NewDecoder(resp.Body).Decode(&kResp); e != nil {
		return err("failed to decode response: " + e.Error())
}

	if len(kResp.Results) == 0 {
		return ok("No results found")
}

	r := kResp.Results[0].Data
	return ok(fmt.Sprintf("Title: %s\nURL: %s\nSnippet: %s", r.Title, r.URL, r.Snippet))
}