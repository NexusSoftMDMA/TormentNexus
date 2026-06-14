package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCreateMemo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "apiToken")
	content, _ :=getString(args, "content")
	if content == "" {
		return err("content is required")
}

	body := map[string]string{"content": content}
	if tags := getString(args, "tags"); tags != "" {
		body["tags"] = tags
	}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://flomoapp.com/api/v1/memo/create", bytes.NewReader(jsonBody))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	return ok("memo created")
}