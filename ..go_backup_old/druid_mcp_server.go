package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleDruidQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	query, _ :=getString(args, "query")
	if url == "" || query == "" {
		return err("url and query are required")
	}
	body, _ := json.Marshal(map[string]string{"query": query})
	req, e := http.NewRequestWithContext(ctx, "POST", url+"/druid/v2/sql", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
	}
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("druid returned status %d: %s", resp.StatusCode, string(respBody)))
	}
	var result interface{}
	if e := json.Unmarshal(respBody, &result); e != nil {
		return err(fmt.Sprintf("failed to parse response: %v", e))
	}
	return ok(fmt.Sprintf("Query result: %v", result))
}

func HandleDruidStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url+"/status/health", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("health check failed: %v", e))
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return success("Druid is healthy")
	}
	return err(fmt.Sprintf("Druid returned status %d", resp.StatusCode))
}