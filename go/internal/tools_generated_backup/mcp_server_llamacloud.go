package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleListModels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.llamacloud.com/v1/models", nil)
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Failed to make request: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleGenerate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("Missing required parameter 'prompt'")
}

	payload := map[string]string{"prompt": prompt}
	bodyBytes, e := json.Marshal(payload)
	if e != nil {
		return err("Failed to marshal payload: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.llamacloud.com/v1/generate", bytes.NewReader(bodyBytes))
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Failed to make request: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	return ok(string(body))
}