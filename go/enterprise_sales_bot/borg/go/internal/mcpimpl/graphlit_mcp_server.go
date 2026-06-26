package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleQuery_graphlit_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	payload := map[string]string{"query": query}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal query")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.graphlit.com/v1/query", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(respBody))
}

func HandleCreate_graphlit_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	payload := map[string]string{"content": content}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal content")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.graphlit.com/v1/contents", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(respBody))
}