package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

func ListCampaigns(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL := os.Getenv("ATRACKER_URL")
	if baseURL == "" {
		return err("ATRACKER_URL not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/campaigns", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if resp.StatusCode >= 400 {
		return err("API error: " + string(body))
}

	return ok(string(body))
}