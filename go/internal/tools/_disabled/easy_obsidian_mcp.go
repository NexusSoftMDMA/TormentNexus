package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchNotes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	encodedQuery := url.QueryEscape(query)
	urlStr := fmt.Sprintf("http://localhost:27123/search?query=%s", encodedQuery)
	req, e := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

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
		return err(fmt.Sprintf("Obsidian API returned status %d: %s", resp.StatusCode, string(body)))
}

	var results []map[string]interface{}
	if e := json.Unmarshal(body, &results); e != nil {
		return err(fmt.Sprintf("failed to parse response: %v", e))
}

	var output string
	for i, r := range results {
		title, found := r["title"].(string)
		if !found {
			title = "unknown"
		}
		path, found := r["path"].(string)
		if !found {
			path = "unknown"
		}
		output += fmt.Sprintf("%d. %s (path: %s)\n", i+1, title, path)

	if output == "" {
		output = "No results found."
	}
	return ok(output)
}

}

func HandleOpenNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	urlStr := fmt.Sprintf("http://localhost:27123/open/%s", url.PathEscape(path))
	req, e := http.NewRequestWithContext(ctx, "POST", urlStr, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("Obsidian API returned status %d", resp.StatusCode))
}

	return ok("Opened note: " + path)
}