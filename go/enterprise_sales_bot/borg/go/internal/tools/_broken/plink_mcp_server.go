package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleCreateShortLink(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	originalURL, _ :=getString(args, "url")
	if originalURL == "" {
		return err("missing required argument: url")
}

	slug, _ :=getString(args, "slug")
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("missing required argument: api_key")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.p-link.io/v1/links", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	body := map[string]string{"url": originalURL}
	if slug != "" {
		body["slug"] = slug
	}
	jsonBody, _ := json.Marshal(body)
	req.Body = io.NopCloser(bytes.NewReader(jsonBody))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return success(fmt.Sprintf("Short link created: %v", result["short_url"]))
}

func HandleGetLinkStats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	linkID, _ :=getString(args, "link_id")
	if linkID == "" {
		return err("missing required argument: link_id")
}

	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("missing required argument: api_key")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.p-link.io/v1/links/%s/stats", linkID), nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return success(fmt.Sprintf("Stats: clicks=%v, created=%v", result["clicks"], result["created_at"]))
}