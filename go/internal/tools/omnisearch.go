//go:build ignore
// +build ignore

package tools

/**
 * @file omnisearch.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Omnisearch MCP — universal search across sources.
 * Replaces: github.com/spences10/mcp-omnisearch
 *
 * Provides unified search across GitHub, Stack Overflow, npm, PyPI, DuckDuckGo,
 * and Wikipedia from a single interface.
 *
 * Tools:
 *  - omnisearch_github — search GitHub repositories
 *  - omnisearch_stackoverflow — search Stack Overflow
 *  - omnisearch_npm — search npm packages
 *  - omnisearch_pypi — search PyPI packages
 *  - omnisearch_web — general web search
 */

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

func ghToken() string { return os.Getenv("GITHUB_TOKEN") }

func newHTTPClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

// HandleOmnisearchGithub searches GitHub repositories.
func HandleOmnisearchGithub(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q")
	if query == "" {
		return err("query is required")
	}
	limit := getInt(args, "limit")
	if limit <= 0 || limit > 50 {
		limit = 5
	}

	client := newHTTPClient()
	req, e := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("https://api.github.com/search/repositories?q=%s&per_page=%d&sort=stars", url.QueryEscape(query), limit), nil)
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if token := ghToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, e := client.Do(req)
	if e != nil {
		return err(fmt.Sprintf("GitHub search failed: %v", e))
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(string(data))
}

// HandleOmnisearchStackoverflow searches Stack Overflow.
func HandleOmnisearchStackoverflow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q")
	if query == "" {
		return err("query is required")
	}
	limit := getInt(args, "limit")
	if limit <= 0 || limit > 50 {
		limit = 5
	}

	client := newHTTPClient()
	resp, e := client.Get(fmt.Sprintf(
		"https://api.stackexchange.com/2.3/search?order=desc&sort=votes&intitle=%s&site=stackoverflow&pagesize=%d",
		url.QueryEscape(query), limit))
	if e != nil {
		return err(fmt.Sprintf("Stack Overflow search failed: %v", e))
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(string(data))
}

// HandleOmnisearchNpm searches npm packages.
func HandleOmnisearchNpm(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q")
	if query == "" {
		return err("query is required")
	}
	limit := getInt(args, "limit")
	if limit <= 0 || limit > 50 {
		limit = 5
	}

	client := newHTTPClient()
	resp, e := client.Get(fmt.Sprintf(
		"https://registry.npmjs.org/-/v1/search?text=%s&size=%d",
		url.QueryEscape(query), limit))
	if e != nil {
		return err(fmt.Sprintf("npm search failed: %v", e))
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(string(data))
}

// HandleOmnisearchPypi searches PyPI packages.
func HandleOmnisearchPypi(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q")
	if query == "" {
		return err("query is required")
	}

	client := newHTTPClient()
	resp, e := client.Get(fmt.Sprintf(
		"https://pypi.org/simple/?q=%s", url.QueryEscape(query)))
	if e != nil {
		return err(fmt.Sprintf("PyPI search failed: %v", e))
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	// Also get detailed JSON
	resp2, e2 := client.Get(fmt.Sprintf(
		"https://pypi.org/pypi/%s/json", query))
	if e2 == nil {
		defer resp2.Body.Close()
		detail, _ := io.ReadAll(resp2.Body)
		result := map[string]interface{}{
			"search_html": string(data),
			"details":     json.RawMessage(detail),
		}
		out, _ := json.MarshalIndent(result, "", "  ")
		return ok(string(out))
	}

	return ok(string(data))
}

// HandleOmnisearchWeb performs general web search via DuckDuckGo.
func HandleOmnisearchWeb(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q")
	if query == "" {
		return err("query is required")
	}

	client := newHTTPClient()
	resp, e := client.Get(fmt.Sprintf(
		"https://api.duckduckgo.com/?q=%s&format=json&no_html=1",
		url.QueryEscape(query)))
	if e != nil {
		return err(fmt.Sprintf("web search failed: %v", e))
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(string(data))
}
