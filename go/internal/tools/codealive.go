//go:build ignore
// +build ignore

package tools

/**
 * @file codealive.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of CodeAlive MCP — semantic code search and context engine.
 * Replaces: io.github.codealive-ai.codealive-mcp
 *
 * Provides semantic code search across indexed repositories using CodeAlive's GraphRAG engine.
 *
 * Tools:
 *  - codealive_search — semantic search across codebase
 *  - codealive_grep — exact text/regex search
 *  - codealive_artifacts — fetch full source for search hits
 *  - codealive_relationships — get call graph / references
 *  - codealive_ask — synthesized codebase Q&A
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

const codealiveDefaultBaseURL = "https://api.codealive.ai/v1"

func codealiveBaseURL() string {
	if u := os.Getenv("CODEALIVE_BASE_URL"); u != "" {
		return u
	}
	return codealiveDefaultBaseURL
}

func codealiveAPIKey() string {
	return os.Getenv("CODEALIVE_API_KEY")
}

func codealivePost(ctx context.Context, path string, payload map[string]interface{}) (string, error) {
	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 60 * time.Second}
	req, e := http.NewRequestWithContext(ctx, "POST",
		codealiveBaseURL()+path, bytes.NewReader(body))
	if e != nil {
		return "", fmt.Errorf("request error: %v", e)
	}
	req.Header.Set("Content-Type", "application/json")
	if key := codealiveAPIKey(); key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	resp, e := client.Do(req)
	if e != nil {
		return "", fmt.Errorf("request failed: %v", e)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(data))
	}
	return string(data), nil
}

// HandleCodeAliveSearch performs semantic code search.
// Required: query (string) — the search query
// Optional: repository, workspace, limit
func HandleCodeAliveSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q")
	if query == "" {
		return err("query is required")
	}
	payload := map[string]interface{}{"query": query}
	if repo, ok := getString(args, "repository", "repo"); ok {
		payload["repository"] = repo
	}
	if limit := getInt(args, "limit"); limit > 0 {
		payload["limit"] = limit
	}

	result, e := codealivePost(ctx, "/search", payload)
	if e != nil {
		return err(fmt.Sprintf("search error: %v", e))
	}
	return ok(result)
}

// HandleCodeAliveGrep performs exact text or regex search.
// Required: pattern (string) — text or regex pattern
// Optional: repository, case_sensitive
func HandleCodeAliveGrep(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pattern, _ := getString(args, "pattern", "query")
	if pattern == "" {
		return err("pattern is required")
	}
	payload := map[string]interface{}{"pattern": pattern}
	if repo, ok := getString(args, "repository", "repo"); ok {
		payload["repository"] = repo
	}
	payload["case_sensitive"] = getBool(args, "case_sensitive", "caseSensitive")

	result, e := codealivePost(ctx, "/grep", payload)
	if e != nil {
		return err(fmt.Sprintf("grep error: %v", e))
	}
	return ok(result)
}

// HandleCodeAliveAsk performs a synthesized codebase Q&A.
// Required: question (string) — the question about the codebase
// Optional: repository
func HandleCodeAliveAsk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	question, _ := getString(args, "question", "query")
	if question == "" {
		return err("question is required")
	}
	payload := map[string]interface{}{"question": question}
	if repo, ok := getString(args, "repository", "repo"); ok {
		payload["repository"] = repo
	}

	result, e := codealivePost(ctx, "/ask", payload)
	if e != nil {
		return err(fmt.Sprintf("ask error: %v", e))
	}
	return ok(result)
}
