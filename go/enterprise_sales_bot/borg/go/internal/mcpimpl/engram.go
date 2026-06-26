package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func HandleSearchMemory_engram(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	baseURL := "http://localhost:8080"
	apiURL := fmt.Sprintf("%s/search?q=%s", baseURL, url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Results []string `json:"results"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	if len(result.Results) == 0 {
		return ok("no memories found")
}

	return success(strings.Join(result.Results, ", "))
}

func HandleAddMemory_engram(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	if content == "" {
		return err("content is required")
}

	baseURL := "http://localhost:8080"
	apiURL := fmt.Sprintf("%s/add?content=%s", baseURL, url.QueryEscape(content))
	resp, e := http.DefaultClient.Post(apiURL, "application/json", nil)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("add failed")
}

	return ok("memory added")
}