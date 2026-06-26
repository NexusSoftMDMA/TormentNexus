package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleListRoutes_higress_ops_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "baseUrl")
	if base == "" {
		return err("baseUrl is required")
}

	url := base + "/v1/routes"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	return ok(string(body))
}

func HandleListUpstreams(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "baseUrl")
	if base == "" {
		return err("baseUrl is required")
}

	url := base + "/v1/upstreams"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	return ok(string(body))
}