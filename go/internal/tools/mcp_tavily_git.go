//go:build ignore
// +build ignore

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

// HandleTavilySearch performs a search using Tavily API.
// Args: "query" (required), "api_key" (optional, falls back to TAVILY_API_KEY env)
func HandleTavilySearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
	}
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		apiKey = os.Getenv("TAVILY_API_KEY")

	if apiKey == "" {
		return err("TAVILY_API_KEY not set")
	}

	body, _ := json.Marshal(map[string]string{
		"api_key": apiKey,
		"query":   query,
	})
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.tavily.com/search", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err("API returned status " + resp.Status)
	}

	raw, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response: " + e.Error())
	}
	return success(string(raw))
}
}
