package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleSemanticSearch_semanticapi_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 5
	}
	apiURL := os.Getenv("SEMANTIC_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8000/search"
	}
	body := map[string]interface{}{"query": query, "limit": limit}
	payload, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request")
}

	req, e := http.NewRequestWithContext(ctx, "POST", apiURL, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Body = http.NoBody
	req.Body = http.MaxBytesReader(nil, nil, 0)
	req.ContentLength = int64(len(payload))
	req.GetBody = func() (interface{}, error) { return nil, nil }
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	jsonResult, _ := json.Marshal(result)
	return ok(string(jsonResult))
}

func HandleSemanticEmbed(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	apiURL := os.Getenv("SEMANTIC_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8000/embed"
	}
	body := map[string]interface{}{"text": text}
	payload, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request")
}

	req, e := http.NewRequestWithContext(ctx, "POST", apiURL, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Body = http.NoBody
	req.Body = http.MaxBytesReader(nil, nil, 0)
	req.ContentLength = int64(len(payload))
	req.GetBody = func() (interface{}, error) { return nil, nil }
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	jsonResult, _ := json.Marshal(result)
	return ok(string(jsonResult))
}