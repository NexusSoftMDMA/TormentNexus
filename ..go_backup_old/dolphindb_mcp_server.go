package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleExecuteScript(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	script, _ :=getString(args, "script")
	if script == "" {
		return err("script parameter required")
	}
	body := fmt.Sprintf(`{"script":"%s"}`, script)
	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8848/run", strings.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read response failed: %v", e))
	}
	var result map[string]interface{}
	if e := json.Unmarshal(data, &result); e != nil {
		return err(fmt.Sprintf("json decode failed: %v", e))
	}
	return ok(fmt.Sprintf("Result: %v", result))
}

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter required")
	}
	body := fmt.Sprintf(`{"query":"%s"}`, query)
	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8848/query", strings.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read response failed: %v", e))
	}
	var result map[string]interface{}
	if e := json.Unmarshal(data, &result); e != nil {
		return err(fmt.Sprintf("json decode failed: %v", e))
	}
	return ok(fmt.Sprintf("Query result: %v", result))
}