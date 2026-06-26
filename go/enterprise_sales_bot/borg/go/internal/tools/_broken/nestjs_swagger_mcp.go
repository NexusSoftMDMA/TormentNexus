package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetEndpoints(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	specURL, _ :=getString(args, "spec_url")
	if specURL == "" {
		return err("spec_url is required")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, specURL, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch spec: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var spec map[string]interface{}
	if e := json.Unmarshal(body, &spec); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	paths, found := spec["paths"].(map[string]interface{})
	if !found {
		return err("no paths found in spec")
}

	var endpoints []string
	for path, methods := range paths {
		methodsMap, found := methods.(map[string]interface{})
		if !found {
			continue
		}
		for method := range methodsMap {
			endpoints = append(endpoints, fmt.Sprintf("%s %s", method, path))

	}
	return ok(fmt.Sprintf("Endpoints:\n%s", joinStrings(endpoints, "\n")))
}

}

func HandleCallEndpoint(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	path, _ :=getString(args, "path")
	method, _ :=getString(args, "method")
	if baseURL == "" || path == "" || method == "" {
		return err("base_url, path, and method are required")
}

	url := baseURL + path
	req, e := http.NewRequestWithContext(ctx, method, url, nil)
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
		return err("failed to read response: " + e.Error())
}

	return ok(fmt.Sprintf("Status: %d\nBody: %s", resp.StatusCode, string(body)))
}

func joinStrings_nestjs_swagger_mcp(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}