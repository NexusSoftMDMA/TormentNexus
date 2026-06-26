package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleCallAiGateway(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	endpoint, _ :=getString(args, "endpoint")
	if endpoint == "" {
		return err("endpoint is required")
}

	body := map[string]string{"prompt": prompt}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal body: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(jsonBody))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

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

	if resp.StatusCode != http.StatusOK {
		return err("API returned status " + resp.Status + ": " + string(respBody))
}

	return ok(string(respBody))
}