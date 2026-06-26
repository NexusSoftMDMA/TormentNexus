package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetPage_growi_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	pagePath, _ :=getString(args, "page_path")
	if baseURL == "" || pagePath == "" {
		return err("base_url and page_path are required")
}

	url := fmt.Sprintf("%s/_api/pages/%s", baseURL, pagePath)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	content, found := result["content"].(string)
	if !found {
		return err("no content in response")
}

	return ok(content)
}

func HandleSearchPages(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	query, _ :=getString(args, "query")
	if baseURL == "" || query == "" {
		return err("base_url and query are required")
}

	url := fmt.Sprintf("%s/_api/search?q=%s", baseURL, query)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	pages, found := result["pages"].([]interface{})
	if !found {
		return err("no pages in response")
}

	data, e := json.Marshal(pages)
	if e != nil {
		return err("failed to marshal pages")
}

	return ok(string(data))
}