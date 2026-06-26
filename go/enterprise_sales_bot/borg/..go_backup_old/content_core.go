package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetContent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing id")
}

	url := fmt.Sprintf("https://api.example.com/content/%s", id)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status")
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error")
}

	return ok(fmt.Sprintf("Content: %v", result))
}

func HandleListContent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	category, _ :=getString(args, "category")
	url := "https://api.example.com/content"
	if category != "" {
		url += "?category=" + category
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status")
}

	var results []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&results); e != nil {
		return err("decode error")
}

	return ok(fmt.Sprintf("Found %d items", len(results)))
}