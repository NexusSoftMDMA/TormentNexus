package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	payload := map[string]string{"query": query}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal: " + e.Error())
}

	resp, e := http.DefaultClient.Post("http://localhost:8080/query", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(respBody))
}

func HandleInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Cortex Plugin MCP server v1.0")
}