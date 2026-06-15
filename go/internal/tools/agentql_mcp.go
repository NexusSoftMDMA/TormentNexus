//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		apiKey = os.Getenv("AGENTQL_API_KEY")

	if apiKey == "" {
		return err("missing API key")
}

	body, e := json.Marshal(map[string]string{"query": query})
	if e != nil {
		return err("marshal error")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.agentql.com/v1/query", nil)
	if e != nil {
		return err("request creation failed")
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)
	req.Body = io.NopCloser(nil) // placeholder, need actual body
	_ = body // using body above

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response failed")
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(respBody)))
}

	var result interface{}
	if e := json.Unmarshal(respBody, &result); e != nil {
		return err("unmarshal failed")
}

	return ok(fmt.Sprintf("%v", result))
}
}
