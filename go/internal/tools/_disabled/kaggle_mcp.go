package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleSearchDatasets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://www.kaggle.com/api/v1/datasets/list", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	q := req.URL.Query()
	q.Set("search", query)
	req.URL.RawQuery = q.Encode()
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Datasets: %v", result))
}

func HandleListKernels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://www.kaggle.com/api/v1/kernels/list", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	q := req.URL.Query()
	q.Set("search", query)
	req.URL.RawQuery = q.Encode()
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Kernels: %v", result))
}