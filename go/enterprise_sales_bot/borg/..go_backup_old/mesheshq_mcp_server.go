package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleEmitEvent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	eventName, _ :=getString(args, "eventName")
	if eventName == "" {
		return err("eventName is required")
}

	payload, _ :=getString(args, "payload")
	baseURL, _ :=getString(args, "baseUrl")
	if baseURL == "" {
		baseURL = "https://api.meshes.sh"
	}
	body, e := json.Marshal(map[string]string{"event": eventName, "payload": payload})
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", baseURL+"/events", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http error: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("bad status: %d", resp.StatusCode))
}

	return success("event emitted")
}

func HandleListWorkspaces(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "baseUrl")
	if baseURL == "" {
		baseURL = "https://api.meshes.sh"
	}
	limit, _ :=getInt(args, "limit")
	url := baseURL + "/workspaces"
	if limit > 0 {
		url = fmt.Sprintf("%s?limit=%d", url, limit)

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http error: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("bad status: %d", resp.StatusCode))
}

	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	return ok(result)
}
}