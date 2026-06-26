package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListFlows_mcp_prefect(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		host = "http://localhost:4200"
	}
	apiKey, _ :=getString(args, "api_key")
	url := host + "/api/flows"
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
	}
	if apiKey != "" {
		req.Header.Set("X-Prefect-Api-Key", apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
	}
	return ok("flows data")
}

}

func HandleListRuns_mcp_prefect(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		host = "http://localhost:4200"
	}
	apiKey, _ :=getString(args, "api_key")
	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 10
	}
	url := fmt.Sprintf("%s/api/runs?limit=%d", host, limit)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
	}
	if apiKey != "" {
		req.Header.Set("X-Prefect-Api-Key", apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
	}
	return ok("runs data")
}
}