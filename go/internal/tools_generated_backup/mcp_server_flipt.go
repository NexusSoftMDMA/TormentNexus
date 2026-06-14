package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListFlags(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL := os.Getenv("FLIPT_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/v1/flags", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	var parsed interface{}
	if e := json.Unmarshal(body, &parsed); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return ok(string(body))
}

func HandleEvaluateFlag(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	flagKey, _ :=getString(args, "flagKey")
	entityId, _ :=getString(args, "entityId")
	contextMap, found := args["context"].(map[string]interface{})
	if !found {
		contextMap = make(map[string]interface{})

	body, e := json.Marshal(map[string]interface{}{
		"flag_key": flagKey,
		"entity_id": entityId,
		"context": contextMap,
	})
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	baseURL := os.Getenv("FLIPT_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	req, e := http.NewRequestWithContext(ctx, "POST", baseURL+"/api/v1/evaluate", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	// Content-Length is set automatically by http.NewRequest with body
	req.Body = io.NopCloser(bytes.NewReader(body)) // need bytes import
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(respBody)))
}

	return ok(string(respBody))
}
}