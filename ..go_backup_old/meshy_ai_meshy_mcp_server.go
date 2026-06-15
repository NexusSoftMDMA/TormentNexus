package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGenerate3D(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
	}
	apiKey := os.Getenv("MESHY_API_KEY")
	if apiKey == "" {
		return err("MESHY_API_KEY not set")
	}
	body := map[string]interface{}{
		"prompt": prompt,
		"style":  getString(args, "style"),
	}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal: %v", e))
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.meshy.ai/v1/text-to-3d", bytes.NewReader(jsonBody))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read response: %v", e))
	}
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(respBody)))
	}
	var result map[string]interface{}
	if e := json.Unmarshal(respBody, &result); e != nil {
		return err(fmt.Sprintf("parse response: %v", e))
	}
	return ok(fmt.Sprintf("Model created: %v", result["id"]))
}