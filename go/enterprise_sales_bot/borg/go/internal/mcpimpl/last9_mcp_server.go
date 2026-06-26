package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleGetTraces(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "api_base")
	service, _ :=getString(args, "service")
	if base == "" || service == "" {
		return err("api_base and service are required")
}

	url := base + "/api/v1/traces?service=" + service
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch traces: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleListServices_last9_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "api_base")
	if base == "" {
		return err("api_base is required")
}

	url := base + "/api/v1/services"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to list services: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}