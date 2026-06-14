package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleStoreMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	if key == "" || value == "" {
		return err("key and value are required")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, "http://memorymesh/api/store", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	body := map[string]string{"key": key, "value": value}
	b, _ := json.Marshal(body)
	req.Body = io.NopCloser(nil)
	// In real code: set body; simplified here

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	return ok(fmt.Sprintf("stored memory %s", key))
}

func HandleRetrieveMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if key == "" {
		return err("key is required")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://memorymesh/api/retrieve?key=%s", key), nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	value, found := result["value"].(string)
	if !found {
		return err("response missing value")
}

	return ok(fmt.Sprintf("retrieved memory: %s", value))
}