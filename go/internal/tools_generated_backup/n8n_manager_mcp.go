package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleListWorkflows(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		return err("base_url is required")
}

	url := baseURL + "/api/v1/workflows"
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("request build error: " + e.Error())
}

	apiKey, _ :=getString(args, "api_key")
	if apiKey != "" {
		req.Header.Set("X-N8N-API-KEY", apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response error: " + e.Error())
}

	return ok(string(body))
}

}

func HandleGetWorkflow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		return err("base_url is required")
}

	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	url := baseURL + "/api/v1/workflows/" + id
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("request build error: " + e.Error())
}

	apiKey, _ :=getString(args, "api_key")
	if apiKey != "" {
		req.Header.Set("X-N8N-API-KEY", apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response error: " + e.Error())
}

	return ok(string(body))
}
}