package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleCheckUrl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing 'url' argument")
}

	apiKey := os.Getenv("VT_API_KEY")
	if apiKey == "" {
		return err("VT_API_KEY environment variable not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://www.virustotal.com/api/v3/urls?url=%s", url), nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("x-apikey", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(string(body))
}

func HandleCheckHash(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	hash, _ :=getString(args, "hash")
	if hash == "" {
		return err("missing 'hash' argument")
}

	apiKey := os.Getenv("VT_API_KEY")
	if apiKey == "" {
		return err("VT_API_KEY environment variable not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://www.virustotal.com/api/v3/files/%s", hash), nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("x-apikey", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return ok(string(body))
}