package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleHttpSseConnect(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url argument is required")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
	}
	req.Header.Set("Accept", "text/event-stream")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("connection failed: %v", e))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status: %s", resp.Status))
	}
	return success(fmt.Sprintf("connected to %s with status %d", url, resp.StatusCode))
}

func HandleHttpJsonPost(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	payload, found := args["payload"]
	if !found {
		return err("payload argument is required")
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal payload: %v", e))
	}
	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(body)))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	return success(fmt.Sprintf("posted to %s, status: %d", url, resp.StatusCode))
}