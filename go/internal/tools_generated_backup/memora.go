package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleListMemories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "http://localhost:8080/memories"
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
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

	return ok(string(body))
}

func HandleAddMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	payload := map[string]string{"content": content}
	bodyBytes, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	url := "http://localhost:8080/memories"
	req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Body = io.NopCloser(io.ReaderFrom(bytes.NewReader(bodyBytes)))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}