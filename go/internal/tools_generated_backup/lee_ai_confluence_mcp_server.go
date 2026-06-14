package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleGetPage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pageID, _ :=getString(args, "page_id")
	baseURL, _ :=getString(args, "base_url")
	token, _ :=getString(args, "api_token")
	if pageID == "" || baseURL == "" {
		return err("page_id and base_url are required")
}

	url := strings.TrimRight(baseURL, "/") + "/rest/api/content/" + pageID
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer " + token)
	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Page: %v", result))
}

func HandleSearchContent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	baseURL, _ :=getString(args, "base_url")
	token, _ :=getString(args, "api_token")
	if query == "" || baseURL == "" {
		return err("query and base_url are required")
}

	url := strings.TrimRight(baseURL, "/") + "/rest/api/search?cql=" + query
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer " + token)
	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(fmt.Sprintf("Search results: %s", string(body)))
}