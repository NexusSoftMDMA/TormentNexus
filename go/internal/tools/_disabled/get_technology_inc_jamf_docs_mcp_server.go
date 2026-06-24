package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchJamfDocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "query")
	if q == "" {
		return err("query parameter is required")
}

	resp, e := http.DefaultClient.Get("https://learn.jamf.com/api/docs/search?q=" + url.QueryEscape(q))
	if e != nil {
		return err(fmt.Sprintf("search failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("search returned status %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read response failed: %v", e))
}

	var results []map[string]any
	if e := json.Unmarshal(body, &results); e != nil {
		return err(fmt.Sprintf("parse search results failed: %v", e))
}

	return ok(fmt.Sprintf("Found %d results", len(results)))
}

func HandleGetJamfDoc(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id parameter is required")
}

	resp, e := http.DefaultClient.Get("https://learn.jamf.com/api/docs/" + url.PathEscape(id))
	if e != nil {
		return err(fmt.Sprintf("fetch doc failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("doc request returned status %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read response failed: %v", e))
}

	var doc map[string]any
	if e := json.Unmarshal(body, &doc); e != nil {
		return err(fmt.Sprintf("parse doc failed: %v", e))
}

	return success(string(body))
}