package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleWebSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("search failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result struct {
		AbstractText string `json:"AbstractText"`
	}
	e = json.Unmarshal(body, &result)
	if e != nil {
		return err("parse failed: " + e.Error())
}

	if result.AbstractText == "" {
		return ok("No results found for " + query)
}

	return ok(result.AbstractText)
}