package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleExecuteQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	query, _ :=getString(args, "query")
	if url == "" || query == "" {
		return err("url and query are required")
}

	body, _ := json.Marshal(map[string]string{"query": query})
	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("failed to decode response: %v", e))
}

	return ok(fmt.Sprintf("GraphQL result: %v", result))
}

func HandleListSchemas(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	introspectionQuery := `{ __schema { types { name } } }`
	body, _ := json.Marshal(map[string]string{"query": introspectionQuery})
	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("failed to decode response: %v", e))
}

	return ok(fmt.Sprintf("Schema list: %v", result))
}