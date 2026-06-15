package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleCreateTouchpoint(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	payload, _ :=getString(args, "payload")
	if url == "" || payload == "" {
		return err("url and payload are required")
}

	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBufferString(payload))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return err("API returned status " + resp.Status)
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

func HandleGetTouchpoint(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	id, _ :=getString(args, "id")
	if url == "" || id == "" {
		return err("url and id are required")
}

	fullURL := url + "/" + id
	req, e := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return err("API returned status " + resp.Status)
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}