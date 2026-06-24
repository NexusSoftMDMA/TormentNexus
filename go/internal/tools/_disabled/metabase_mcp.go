package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListDashboards(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	key, _ :=getString(args, "api_key")
	if host == "" || key == "" {
		return err("host and api_key required")
}

	url := fmt.Sprintf("%s/api/dashboard", host)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("X-API-KEY", key)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Dashboards: %v", result))
}

func HandleRunQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	key, _ :=getString(args, "api_key")
	query, _ :=getString(args, "query")
	if host == "" || key == "" || query == "" {
		return err("host, api_key, and query required")
}

	url := fmt.Sprintf("%s/api/dataset", host)
	body := map[string]string{"query": query}
	payload, e := json.Marshal(body)
	if e != nil {
		return err("marshal failed: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("X-API-KEY", key)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Query result: %v", result))
}