package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleMoire(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pattern, _ :=getString(args, "pattern")
	if pattern == "" {
		pattern = "default"
	}
	url := "https://api.example.com/moire?pattern=" + pattern
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
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("decode failed: " + e.Error())
}

	data, e := json.Marshal(result)
	if e != nil {
		return err("marshal failed: " + e.Error())
}

	return ok(string(data))
}

func HandleMoireStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("M Moire server is active")
}// touch 1781132130
