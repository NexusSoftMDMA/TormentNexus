package mcpimpl

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleGetTrends_trendradar(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyword, _ :=getString(args, "keyword")
	url := "https://api.trendradar.example.com/trends"
	if keyword != "" {
		url += "?keyword=" + keyword
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + url)
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body")
	}
	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response")
	}
	return success("Trends: " + string(body))
}