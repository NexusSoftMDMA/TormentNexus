package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func HandleShareMd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	if content == "" {
		return err("content is required")
}

	apiURL := "https://mdshare.com/api/create"
	resp, e := http.DefaultClient.Post(apiURL, "text/plain", strings.NewReader(content))
	if e != nil {
		return err(fmt.Sprintf("failed to create share: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		ID string `json:"id"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse error: %v", e))
}

	return ok(fmt.Sprintf("Shared md. ID: %s", result.ID))
}

func HandleGetMd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	apiURL := fmt.Sprintf("https://mdshare.com/api/get/%s", url.PathEscape(id))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	body, _ := io.ReadAll(resp.Body)
	return ok(string(body))
}