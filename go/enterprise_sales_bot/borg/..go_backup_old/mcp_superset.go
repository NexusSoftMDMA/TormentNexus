package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListDashboards(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "http://localhost:8088"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/dashboard/", baseURL), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
	}
	return success(fmt.Sprintf("Dashboards: %v", result))
}

func HandleGetDataset(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	id, _ :=getString(args, "id")
	if baseURL == "" || id == "" {
		return err("missing base_url or id")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/dataset/%s", baseURL, id), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
	}
	return success(fmt.Sprintf("Dataset %s: %v", id, result))
}// touch 1781132134
