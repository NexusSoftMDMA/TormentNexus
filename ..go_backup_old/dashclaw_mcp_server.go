package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGuard(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing url")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error: " + e.Error())
}

	blocked, found := result["blocked"]
	if !found {
		return err("missing 'blocked' in response")
}

	if blocked.(bool) {
		return success("action blocked")
}

	return ok("action allowed")
}

func HandleRecord(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing url")
}

	action, _ :=getString(args, "action")
	if action == "" {
		return err("missing action")
}

	req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
	if e != nil {
		return err("request create error: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	return success("action recorded")
}