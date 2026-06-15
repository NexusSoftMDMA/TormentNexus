//go:build ignore
// +build ignore

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleChatCompletions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ :=getString(args, "model")
	if model == "" {
		model = "gpt-4o-mini"
	}
	messages := args["messages"]
	body, e := json.Marshal(map[string]interface{}{
		"model":    model,
		"messages": messages,
	})
	if e != nil {
		return err("failed to marshal request")
}

	endpoint, _ :=getString(args, "endpoint")
	if endpoint == "" {
		endpoint = "https://api..com/v1/chat/completions"
	}
	req, e := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if e != nil {
		return err("failed to create request")
}

	apiKey, _ :=getString(args, "api_key")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(respBody))
}
}
