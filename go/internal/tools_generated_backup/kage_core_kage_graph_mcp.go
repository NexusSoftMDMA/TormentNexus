package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// HandleQuery retrieves code graph nodes matching a query.
func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	baseURL := "http://localhost:8000"
	u, e := url.Parse(baseURL + "/graph/query")
	if e != nil {
		return err("failed to parse url")
	}
	q := u.Query()
	q.Set("q", query)
	q.Set("limit", fmt.Sprintf("%d", limit))
	u.RawQuery = q.Encode()
	resp, e := http.DefaultClient.Get(u.String())
	if e != nil {
		return err("query failed: "+e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: "+e.Error())
	}
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: "+e.Error())
	}
	return ok(fmt.Sprintf("Found %v results", result["count"]))
}