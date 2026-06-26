package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleListVectors(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "api_url")
	if url == "" {
		return err("api_url is required")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return ok(string(body))
}

func HandleGetVector(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "api_url")
	id, _ :=getString(args, "id")
	if url == "" || id == "" {
		return err("api_url and id are required")
}

	fullURL := fmt.Sprintf("%s/%s", url, id)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return ok(string(body))
}