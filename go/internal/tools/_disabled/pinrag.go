package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandlePinragList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// GET request to list items
	resp, e := http.DefaultClient.Get("http://localhost:8080/pins")
	if e != nil {
		return err(fmt.Sprintf("failed to list pins: %v", e))
}

	defer resp.Body.Close()
	var result []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

func HandlePinragAdd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	content, _ :=getString(args, "content")
	if title == "" || content == "" {
		return err("title and content are required")
}

	params := url.Values{}
	params.Set("title", title)
	params.Set("content", content)
	resp, e := http.DefaultClient.PostForm("http://localhost:8080/pins", params)
	if e != nil {
		return err(fmt.Sprintf("failed to add pin: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	return success("pin added successfully")
}