//go:build ignore
// +build ignore

package tools

/**
 * @file webpeel.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of WebPeel — web data API for AI agents.
 * Replaces: github.com/webpeel/webpeel
 *
 * Web Peeling: Extract structured data from any web page via a simple API.
 * Configurable via WEBPEEL_API_KEY env var.
 *
 * Tools:
 *  - webpeel_fetch — fetch and extract content from a URL
 *  - webpeel_search — search the web and extract results
 *  - webpeel_extract — extract structured data using selectors
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

func webpeelAPIKey() string {
	return os.Getenv("WEBPEEL_API_KEY")
}

func webpeelBaseURL() string {
	if u := os.Getenv("WEBPEEL_BASE_URL"); u != "" {
		return u
	}
	return "https://api.webpeel.dev/v1"
}

func webpeelRequest(ctx context.Context, method, path string, payload map[string]interface{}) (string, error) {
	var bodyReader io.Reader
	if payload != nil {
		b, _ := json.Marshal(payload)
		bodyReader = bytes.NewReader(b)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	req, e := http.NewRequestWithContext(ctx, method, webpeelBaseURL()+path, bodyReader)
	if e != nil {
		return "", fmt.Errorf("request error: %v", e)
	}
	req.Header.Set("Content-Type", "application/json")
	if key := webpeelAPIKey(); key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}

	resp, e := client.Do(req)
	if e != nil {
		return "", fmt.Errorf("API request failed: %v", e)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(data))
	}
	return string(data), nil
}

// HandleWebpeelFetch fetches and extracts content from a URL.
func HandleWebpeelFetch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ := getString(args, "url", "target")
	if targetURL == "" {
		return err("url is required")
	}

	payload := map[string]interface{}{"url": targetURL}
	if format, _ := getString(args, "format", "output"); format != "" {
		payload["format"] = format
	}

	result, e := webpeelRequest(ctx, "POST", "/fetch", payload)
	if e != nil {
		return err(fmt.Sprintf("fetch failed: %v", e))
	}
	return ok(result)
}

// HandleWebpeelSearch searches the web and extracts results.
func HandleWebpeelSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q", "search")
	if query == "" {
		return err("query is required")
	}

	payload := map[string]interface{}{"query": query}
	if limit := getInt(args, "limit"); limit > 0 {
		payload["limit"] = limit
	}

	result, e := webpeelRequest(ctx, "POST", "/search", payload)
	if e != nil {
		return err(fmt.Sprintf("search failed: %v", e))
	}
	return ok(result)
}

// HandleWebpeelExtract extracts structured data from a URL using CSS selectors.
func HandleWebpeelExtract(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ := getString(args, "url", "target")
	if targetURL == "" {
		return err("url is required")
	}
	selector, _ := getString(args, "selector", "css")
	if selector == "" {
		return err("selector is required")
	}

	payload := map[string]interface{}{
		"url":      targetURL,
		"selector": selector,
	}

	result, e := webpeelRequest(ctx, "POST", "/extract", payload)
	if e != nil {
		return err(fmt.Sprintf("extract failed: %v", e))
	}
	return ok(result)
}
