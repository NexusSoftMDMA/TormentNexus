//go:build ignore
// +build ignore

package tools

/**
 * @file panther_mcp.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Panther MCP — security monitoring.
 * Replaces: github.com/panther-labs/mcp-panther
 *
 * Integrates with Panther security cloud for detection engineering,
 * querying security findings, and managing policies.
 * Configurable via PANTHER_API_URL and PANTHER_API_TOKEN env vars.
 *
 * Tools:
 *  - panther_query — query security data
 *  - panther_list_detections — list detection rules
 *  - panther_get_findings — get security findings
 *  - panther_list_policies — list security policies
 */

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func pantherBaseURL() string {
	if u := os.Getenv("PANTHER_API_URL"); u != "" {
		return u
	}
	return "https://api.panther.com/v1"
}

func pantherToken() string {
	return os.Getenv("PANTHER_API_TOKEN")
}

func pantherRequest(ctx context.Context, method, path string, payload map[string]interface{}) (string, error) {
	var bodyReader io.Reader
	if payload != nil {
		b, _ := json.Marshal(payload)
		bodyReader = bytes.NewReader(b)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, e := http.NewRequestWithContext(ctx, method, pantherBaseURL()+path, bodyReader)
	if e != nil {
		return "", fmt.Errorf("request error: %v", e)
	}
	req.Header.Set("Content-Type", "application/json")
	if token := pantherToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, e := client.Do(req)
	if e != nil {
		return "", fmt.Errorf("Panther API error: %v", e)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("Panther error (%d): %s", resp.StatusCode, string(data))
	}
	return string(data), nil
}

// HandlePantherQuery executes a Panther security query.
func HandlePantherQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q", "sql")
	if query == "" {
		return err("query is required")
	}

	payload := map[string]interface{}{"query": query}
	result, e := pantherRequest(ctx, "POST", "/query", payload)
	if e != nil {
		return err(fmt.Sprintf("query failed: %v", e))
	}
	return ok(result)
}

// HandlePantherListDetections lists detection rules.
func HandlePantherListDetections(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path := "/detections"
	if limit := getInt(args, "limit"); limit > 0 {
		path = fmt.Sprintf("/detections?limit=%d", limit)
	}
	result, e := pantherRequest(ctx, "GET", path, nil)
	if e != nil {
		return err(fmt.Sprintf("list detections failed: %v", e))
	}
	return ok(result)
}

// HandlePantherGetFindings gets security findings.
func HandlePantherGetFindings(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	severity, _ := getString(args, "severity", "sev")
	path := "/findings"
	if severity != "" {
		path += "?severity=" + severity
	}
	result, e := pantherRequest(ctx, "GET", path, nil)
	if e != nil {
		return err(fmt.Sprintf("get findings failed: %v", e))
	}
	return ok(result)
}

// HandlePantherListPolicies lists security policies.
func HandlePantherListPolicies(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path := "/policies"
	if limit := getInt(args, "limit"); limit > 0 {
		path = fmt.Sprintf("/policies?limit=%d", limit)
	}
	result, e := pantherRequest(ctx, "GET", path, nil)
	if e != nil {
		return err(fmt.Sprintf("list policies failed: %v", e))
	}
	return ok(result)
}
