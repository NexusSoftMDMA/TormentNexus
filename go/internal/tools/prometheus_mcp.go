//go:build ignore
// +build ignore

package tools

/**
 * @file prometheus_mcp.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Prometheus MCP server.
 * Replaces: npm prometheus-mcp
 *
 * Provides natural language interaction with Prometheus monitoring infrastructure.
 * Tools: query, alerts, targets, metadata
 */

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const promDefaultURL = "http://localhost:9090"

func promBaseURL() string {
	if u := os.Getenv("PROMETHEUS_URL"); u != "" {
		return u
	}
	return promDefaultURL
}

// HandlePromQuery runs a PromQL instant query.
func HandlePromQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q", "promql")
	if query == "" {
		return err("query is required")
	}

	baseURL := promBaseURL()
	reqURL := fmt.Sprintf("%s/api/v1/query?query=%s", baseURL, url.QueryEscape(query))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Get(reqURL)
	if e != nil {
		return err(fmt.Sprintf("query failed: %v", e))
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("Prometheus error (%d): %s", resp.StatusCode, string(data)))
	}
	return ok(string(data))
}

// HandlePromAlerts lists active Prometheus alerts.
func HandlePromAlerts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Get(promBaseURL() + "/api/v1/alerts")
	if e != nil {
		return err(fmt.Sprintf("alerts failed: %v", e))
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(string(data))
}

// HandlePromTargets lists monitored targets and their state.
func HandlePromTargets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	state, _ := getString(args, "state", "filter")
	path := promBaseURL() + "/api/v1/targets"
	if state != "" {
		path += "?state=" + state
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Get(path)
	if e != nil {
		return err(fmt.Sprintf("targets failed: %v", e))
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(string(data))
}

// HandlePromMetadata returns metric metadata.
func HandlePromMetadata(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	metric, _ := getString(args, "metric", "name")
	path := promBaseURL() + "/api/v1/metadata"
	if metric != "" {
		path += "?metric=" + url.QueryEscape(metric)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Get(path)
	if e != nil {
		return err(fmt.Sprintf("metadata failed: %v", e))
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(string(data))
}
