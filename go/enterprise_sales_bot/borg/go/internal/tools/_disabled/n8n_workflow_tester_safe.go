package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleTestWorkflow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "workflow_url")
	payloadStr, _ :=getString(args, "payload")
	if url == "" {
		return err("workflow_url is required")
}

	var payload interface{}
	if payloadStr != "" {
		if e := json.Unmarshal([]byte(payloadStr), &payload); e != nil {
			return err("invalid payload JSON")

	}
	body, _ := json.Marshal(payload)
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("n8n returned status %d: %s", resp.StatusCode, string(respBody)))
}

	return ok(string(respBody))
}
}