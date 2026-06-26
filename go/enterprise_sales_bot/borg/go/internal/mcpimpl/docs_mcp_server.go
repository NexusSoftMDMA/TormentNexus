package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchDocs_docs_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	u := fmt.Sprintf("https://api.example.com/docs/search?q=%s", url.QueryEscape(query))
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
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
		Results []string `json:"results"`
	}
	if e = json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	text := "Search results:\n"
	for _, r := range result.Results {
		text += "- " + r + "\n"
	}
	return ok(text)
}

func HandleGetDoc_docs_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	docID, _ :=getString(args, "doc_id")
	if docID == "" {
		return err("doc_id is required")
}

	u := fmt.Sprintf("https://api.example.com/docs/%s", url.PathEscape(docID))
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
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

	var doc struct {
		Content string `json:"content"`
	}
	if e = json.Unmarshal(body, &doc); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(doc.Content)
}