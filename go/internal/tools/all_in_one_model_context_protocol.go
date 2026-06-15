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
	q, _ :=getString(args, "query")
	if q == "" {
		return err("query is required")
}

	u := "https://api.duckduckgo.com/?q=" + url.QueryEscape(q) + "&format=json"
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err(fmt.Sprintf("search failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	abstract, found := result["AbstractText"]
	if !found {
		return ok("No abstract found")
}

	return ok(fmt.Sprintf("%v", abstract))
}

func HandleGetIssue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	service, _ :=getString(args, "service")
	id, _ :=getString(args, "id")
	if service == "" || id == "" {
		return err("service and id are required")
}

	return success(fmt.Sprintf("Fetched %s issue %s (mock)", service, id))
}
