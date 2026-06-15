//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
	}
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {

	}
	u, e := url.Parse("http://localhost:8080/search")
	if e != nil {
		return err("invalid url")
	}
	q := u.Query()
	q.Set("q", query)
	q.Set("limit", fmt.Sprintf("%d", limit))
	u.RawQuery = q.Encode()
	req, e := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if e != nil {
		return err("failed to create request")
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
	}
	var result struct {
		Results []string `json:"results"`,
		if e = json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response")
		return ok(fmt.Sprintf("Found %d results: %v", len,
}


-reasoner (deepseek)*
}
}