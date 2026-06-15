package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetDashboards(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "url")
	token, _ :=getString(args, "token")
	if baseURL == "" {
		return err("missing url parameter")
}

	if token == "" {
		return err("missing token parameter")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/dashboard", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("X-API-Key", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read body: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return success(fmt.Sprintf("Dashboard data: %v", result))
}