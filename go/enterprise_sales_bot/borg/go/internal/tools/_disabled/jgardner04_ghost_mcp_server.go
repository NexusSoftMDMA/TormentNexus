package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleGetPosts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	apiKey, _ :=getString(args, "api_key")
	if url == "" || apiKey == "" {
		return err("url and api_key are required")
}

	u := strings.TrimRight(url, "/") + "/ghost/api/v2/admin/posts/"
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Ghost "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	data, e := json.Marshal(result)
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	return ok(string(data))
}

func HandleCreatePost(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	apiKey, _ :=getString(args, "api_key")
	title, _ :=getString(args, "title")
	content, _ :=getString(args, "content")
	if url == "" || apiKey == "" || title == "" {
		return err("url, api_key, and title are required")
}

	u := strings.TrimRight(url, "/") + "/ghost/api/v2/admin/posts/"
	body := map[string]interface{}{
		"posts": []map[string]interface{}{
			{"title": title, "html": content},
		},
	}
	payload, _ := json.Marshal(body)
	req, e := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(payload))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Ghost "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	data, e := json.Marshal(result)
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	return ok(string(data))
}