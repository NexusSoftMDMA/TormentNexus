package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleOpensearchSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	index, _ :=getString(args, "index")
	query, _ :=getString(args, "query")
	size, _ :=getInt(args, "size")
	if size == 0 {
		size = 10
	}

	body := map[string]interface{}{
		"query": map[string]interface{}{
			"query_string": map[string]interface{}{
				"query": query,
			},
		},
		"size": size,
	}
	payload, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request body")
}

	url := fmt.Sprintf("%s/%s/_search", host, index)
	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	return ok(fmt.Sprintf("Search results: %v", result))
}

func HandleOpensearchIndex(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	index, _ :=getString(args, "index")
	doc, _ :=getString(args, "document")
	id, _ :=getString(args, "id")

	url := fmt.Sprintf("%s/%s/_doc", host, index)
	if id != "" {
		url += "/" + id
	}

	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader([]byte(doc)))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	return ok(fmt.Sprintf("Document indexed: %v", result))
}