package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListEndpoints(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "baseURL")
	if baseURL == "" {
		return err("baseURL is required")
}

	resp, e := http.DefaultClient.Get(baseURL + "/v3/api-docs")
	if e != nil {
		return err("failed to fetch swagger doc: " + e.Error())
}

	defer resp.Body.Close()
	var doc map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&doc); e != nil {
		return err("failed to parse swagger doc: " + e.Error())
}

	paths, found := doc["paths"].(map[string]interface{})
	if !found {
		return err("swagger doc has no paths")
}

	var endpoints []string
	for path := range paths {
		endpoints = append(endpoints, path)

	return ok(fmt.Sprintf("Endpoints: %v", endpoints))
}

}

func HandleGetEndpoint(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "baseURL")
	path, _ :=getString(args, "path")
	if baseURL == "" || path == "" {
		return err("baseURL and path are required")
}

	resp, e := http.DefaultClient.Get(baseURL + "/v3/api-docs")
	if e != nil {
		return err("failed to fetch swagger doc: " + e.Error())
}

	defer resp.Body.Close()
	var doc map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&doc); e != nil {
		return err("failed to parse swagger doc: " + e.Error())
}

	paths, found := doc["paths"].(map[string]interface{})
	if !found {
		return err("swagger doc has no paths")
}

	endpoint, found := paths[path].(map[string]interface{})
	if !found {
		return err("path not found in swagger doc")
}

	data, _ := json.MarshalIndent(endpoint, "", "  ")
	return success(string(data))
}