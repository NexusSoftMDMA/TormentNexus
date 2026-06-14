package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchPapers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	u := fmt.Sprintf("https://paperswithcode.com/api/v1/papers/?q=%s", url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result struct {
		Count   int            `json:"count"`
		Results []map[string]interface{} `json:"results"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json error: %v", e))
}

	if result.Count == 0 {
		return ok("No papers found")
}

	first := result.Results[0]
	title, _ := first["title"].(string)
	return ok(fmt.Sprintf("Found %d papers. First: %s", result.Count, title))
}

func HandleGetPaper(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	u := fmt.Sprintf("https://paperswithcode.com/api/v1/papers/%s", url.PathEscape(id))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var paper map[string]interface{}
	if e := json.Unmarshal(body, &paper); e != nil {
		return err(fmt.Sprintf("json error: %v", e))
}

	title, found := paper["title"].(string)
	if !found {
		return err("paper not found")
}

	return ok(fmt.Sprintf("Paper: %s", title))
}