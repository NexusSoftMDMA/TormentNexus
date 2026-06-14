package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	endpoint, _ :=getString(args, "endpoint")
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	u := fmt.Sprintf("%s/api/v1/query?query=%s", endpoint, url.QueryEscape(query))
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
		return err("read body failed: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse response failed: " + e.Error())
}

	return ok(fmt.Sprintf("query result: %v", result))
}

func HandleHealth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	endpoint, _ :=getString(args, "endpoint")
	if endpoint == "" {
		endpoint = "http://localhost:8428"
	}
	u := endpoint + "/health"
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("health check failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return success("VictoriaMetrics is healthy")
}

	return err("health check returned status: " + resp.Status)
}