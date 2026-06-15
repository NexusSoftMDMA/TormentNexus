//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	registryURL, _ :=getString(args, "registry_url")
	if registryURL == "" {
		registryURL = "https://registry.mcp.com/api/servers"
	}
	reqURL := registryURL
	if query != "" {
		reqURL = registryURL + "?q=" + query
	}
	resp, e := http.DefaultClient.Get(reqURL)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result []map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON")
}

	return success(fmt.Sprintf("Found %d servers: %+v", len(result), result))
}

func HandleInspect(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	registryURL, _ :=getString(args, "registry_url")
	if registryURL == "" {
		registryURL = "https://registry.mcp.com/api/servers"
	}
	reqURL := strings.TrimRight(registryURL, "/") + "/" + name
	resp, e := http.DefaultClient.Get(reqURL)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON")
}

	return success(fmt.Sprintf("Server %s: %+v", name, result))
}
