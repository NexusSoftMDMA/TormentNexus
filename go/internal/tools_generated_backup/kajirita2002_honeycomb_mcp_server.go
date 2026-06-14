package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleListDatasets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	if apiKey == "" {
		return err("apiKey is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.honeycomb.io/1/datasets", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("X-Honeycomb-Team", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

func HandleRunQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	dataset, _ :=getString(args, "dataset")
	queryJSON, _ :=getString(args, "query")
	if apiKey == "" || dataset == "" || queryJSON == "" {
		return err("apiKey, dataset, and query are required")
}

	url := fmt.Sprintf("https://api.honeycomb.io/1/query/%s", dataset)
	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(queryJSON))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("X-Honeycomb-Team", apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}