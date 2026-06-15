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

func HandleGraphqlQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	endpoint, _ :=getString(args, "endpoint")
	query, _ :=getString(args, "query")
	if endpoint == "" || query == "" {
		return err("endpoint and query are required")
}

	payload := map[string]string{"query": query}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal request")
}

	req, e := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
	return ok(string(respBody))
}

func HandleGraphqlSchema(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	endpoint, _ :=getString(args, "endpoint")
	if endpoint == "" {
		return err("endpoint is required")
}

	introspectionQuery := `{ __schema { types { name fields { name type { name kind } } } } }`
	return HandleGraphqlQuery(ctx, map[string]interface{}{
		"query":    introspectionQuery,
	}),
}
}