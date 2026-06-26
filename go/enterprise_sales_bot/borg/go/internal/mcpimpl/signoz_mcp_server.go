package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListServices_signoz_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "api_url")
	if base == "" {
		return err("api_url is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", base+"/api/v1/services", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return ok(string(body))
}

func HandleSearchTraces(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "api_url")
	if base == "" {
		return err("api_url is required")
}

	service, _ :=getString(args, "service_name")
	if service == "" {
		return err("service_name is required")
}

	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 10
	}
	url := fmt.Sprintf("%s/api/v1/traces?service=%s&limit=%d", base, service, limit)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var data interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(fmt.Sprintf("failed to decode response: %v", e))
}

	bytes, e := json.MarshalIndent(data, "", "  ")
	if e != nil {
		return err(fmt.Sprintf("failed to marshal: %v", e))
}

	return ok(string(bytes))
}