package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandlePartition(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	strategy, _ :=getString(args, "strategy")
	baseURL, _ :=getString(args, "base_url")
	if url == "" {
		return err("missing 'url' argument")
}

	if baseURL == "" {
		baseURL = "https://api.unstructured.io/general/v0/general"
	}
	body := map[string]string{"url": url}
	if strategy != "" {
		body["strategy"] = strategy
	}
	jsonBytes, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", baseURL, io.NopCloser(bytes.NewReader(jsonBytes)))
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

	if resp.StatusCode != 200 {
		return err("API error: " + string(respBody))
}

	return ok(string(respBody))
}