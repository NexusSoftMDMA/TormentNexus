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
	"os"
	"time"
)

// HandleFirecrawl Scrapes websites and formats to markdown natively using Firecrawl API.
func HandleFirecrawl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	op, _ := getString(args, "operation", "op")
	if op == "" {
		op = "scrape"
	}

	urlStr, _ := getString(args, "url")
	if urlStr == "" {
		return err("url parameter is required")
	}

	apiKey := os.Getenv("FIRECRAWL_API_KEY")
	if apiKey == "" {
		return err("FIRECRAWL_API_KEY environment variable is not set")
	}

	client := &http.Client{
		Timeout: 45 * time.Second,
	}

	switch op {
	case "scrape":
		payload := map[string]interface{}{
			"url": urlStr,
		}
		if formats, ok := args["formats"].([]interface{}); ok {
			payload["formats"] = formats
		} else {
			payload["formats"] = []string{"markdown"}
		}
		if onlyMain, ok := args["onlyMainContent"].(bool); ok {
			payload["onlyMainContent"] = onlyMain
		}

		jsonPayload, _ := json.Marshal(payload)
		req, errReq := http.NewRequestWithContext(ctx, "POST", "https://api.firecrawl.dev/v1/scrape", bytes.NewBuffer(jsonPayload))
		if errReq != nil {
			return err(fmt.Sprintf("Failed to create request: %v", errReq))
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, errDo := client.Do(req)
		if errDo != nil {
			return err(fmt.Sprintf("Scrape request failed: %v", errDo))
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			return err(fmt.Sprintf("Firecrawl API error (HTTP %d): %s", resp.StatusCode, string(body)))
		}
		return ok(string(body))

	case "crawl":
		payload := map[string]interface{}{
			"url": urlStr,
		}
		if limit := getInt(args, "limit"); limit > 0 {
			payload["limit"] = limit
		}
		if maxDepth := getInt(args, "maxDepth"); maxDepth > 0 {
			payload["maxDepth"] = maxDepth
		}

		jsonPayload, _ := json.Marshal(payload)
		req, errReq := http.NewRequestWithContext(ctx, "POST", "https://api.firecrawl.dev/v1/crawl", bytes.NewBuffer(jsonPayload))
		if errReq != nil {
			return err(fmt.Sprintf("Failed to create request: %v", errReq))
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, errDo := client.Do(req)
		if errDo != nil {
			return err(fmt.Sprintf("Crawl request failed: %v", errDo))
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			return err(fmt.Sprintf("Firecrawl API error (HTTP %d): %s", resp.StatusCode, string(body)))
		}
		return ok(string(body))

	default:
		return err("Unsupported Firecrawl operation: " + op)
	}
}
