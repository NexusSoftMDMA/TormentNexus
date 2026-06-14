package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleCreatePost(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	raw, _ :=getString(args, "raw")
	categoryID, _ :=getInt(args, "category_id")
	apiURL, _ :=getString(args, "api_url")
	apiKey, _ :=getString(args, "api_key")
	if title == "" || raw == "" || apiURL == "" || apiKey == "" {
		return err("missing required arguments: title, raw, api_url, api_key")
}

	body := map[string]interface{}{
		"title": title, "raw": raw, "category": categoryID,
	}
	b, e := json.Marshal(body)
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", apiURL+"/posts.json", bytes.NewReader(b))
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", apiKey)
	req.Header.Set("Api-Username", "system")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http error: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("api error %d: %s", resp.StatusCode, string(bodyBytes)))
}

	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	return ok(fmt.Sprintf("post created: %v", result))
}

func HandleListTopics(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiURL, _ :=getString(args, "api_url")
	apiKey, _ :=getString(args, "api_key")
	categoryID, _ :=getInt(args, "category_id")
	if apiURL == "" || apiKey == "" {
		return err("missing required arguments: api_url, api_key")
}

	path := apiURL + "/latest.json"
	if categoryID != 0 {
		path += fmt.Sprintf("?category=%d", categoryID)

	req, e := http.NewRequestWithContext(ctx, "GET", path, nil)
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
}

	req.Header.Set("Api-Key", apiKey)
	req.Header.Set("Api-Username", "system")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http error: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("api error %d: %s", resp.StatusCode, string(bodyBytes)))
}

	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	return ok(fmt.Sprintf("topics: %v", result))
}
}