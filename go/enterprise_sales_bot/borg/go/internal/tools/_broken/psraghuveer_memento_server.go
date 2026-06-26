package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleMementoRequest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	method, _ :=getString(args, "method")
	if method == "" {
		method = "GET"
	}
	bodyStr, _ :=getString(args, "body")
	var body io.Reader
	if bodyStr != "" {
		body = bytes.NewBufferString(bodyStr)

	req, e := http.NewRequestWithContext(ctx, method, url, body)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBytes, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode >= 400 {
		return err("HTTP " + resp.Status + ": " + string(respBytes))
}

	var result interface{}
	json.Unmarshal(respBytes, &result)
	return ok(string(respBytes))
}
}