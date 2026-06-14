package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleReportSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyword, _ :=getString(args, "keyword")
	apiKey, _ :=getString(args, "apiKey")
	baseUrl, _ :=getString(args, "baseUrl")
	if keyword == "" || apiKey == "" || baseUrl == "" {
		return err("missing required parameters")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseUrl+"/reports?search="+keyword, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	return ok(string(body))
}

func HandleIndicatorGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	indicatorId, _ :=getString(args, "indicator_id")
	apiKey, _ :=getString(args, "apiKey")
	baseUrl, _ :=getString(args, "baseUrl")
	if indicatorId == "" || apiKey == "" || baseUrl == "" {
		return err("missing required parameters")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseUrl+"/indicators/"+indicatorId, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	return ok(string(body))
}