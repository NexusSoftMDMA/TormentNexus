package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleListRoutes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		return err("base_url is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/rails/info/routes", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch routes: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	return success(string(body))
}

func HandleGetVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Rails MCP Server v1.0.0")
}