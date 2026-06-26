package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleShorten_filooapp_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://filoo.app/api/shorten", strings.NewReader(fmt.Sprintf(`{"url":"%s"}`, url)))
	if e != nil {
		return err("failed to create request"), e
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed"), e
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response"), e
	}
	var result struct {
		ShortURL string `json:"short_url"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid response"), e
	}
	return ok(fmt.Sprintf("Short URL: %s", result.ShortURL))
}

func HandleGetClicks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "short_code")
	if code == "" {
		return err("short_code is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://filoo.app/api/clicks/%s", code), nil)
	if e != nil {
		return err("failed to create request"), e
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed"), e
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response"), e
	}
	var result struct {
		Clicks int `json:"clicks"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid response"), e
	}
	return ok(fmt.Sprintf("Total clicks: %d", result.Clicks))
}