package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleGetSecret(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	server, _ :=getString(args, "server")
	path, _ :=getString(args, "path")
	key, _ :=getString(args, "key")
	if server == "" || path == "" || key == "" {
		return err("server, path, and key are required")
}

	url := strings.TrimRight(server, "/") + "/v1/" + strings.Trim(path, "/") + "/" + strings.Trim(key, "/")
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

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	return ok(string(body))
}

func HandleSetSecret(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	server, _ :=getString(args, "server")
	path, _ :=getString(args, "path")
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	if server == "" || path == "" || key == "" || value == "" {
		return err("server, path, key, and value are required")
}

	url := strings.TrimRight(server, "/") + "/v1/" + strings.Trim(path, "/") + "/" + strings.Trim(key, "/")
	payload := map[string]string{"value": value}
	jsonPayload, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal payload: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, http.MethodPut, url, strings.NewReader(string(jsonPayload)))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	return success("secret set successfully")
}