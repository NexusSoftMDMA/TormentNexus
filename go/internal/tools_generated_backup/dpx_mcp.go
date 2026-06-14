package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleDpxSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter required")
}

	url := "https://api.dpx.ai/v1/search?q=" + query
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("request creation failed: " + e.Error())
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

	return success("search completed")
}

func HandleDpxStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.dpx.ai/v1/status", nil)
	if e != nil {
		return err("request creation failed: " + e.Error())
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

	return ok("status retrieved")
}// touch 1781132125
