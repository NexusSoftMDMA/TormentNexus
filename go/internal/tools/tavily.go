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

// HandleTavilySearch performs web searches using the Tavily API natively.
func HandleTavilySearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query")
	if query == "" {
		return err("query parameter is required")
	}

	apiKey := os.Getenv("TAVILY_API_KEY")
	if apiKey == "" {
		return err("TAVILY_API_KEY environment variable is not set")
	}

	searchDepth := getStringValue(args, "search_depth")
	if searchDepth == "" {
		searchDepth = "basic"
	}

	// Build payload mapping basic and advanced Tavily parameters
	payload := map[string]interface{}{
		"api_key":      apiKey,
		"query":        query,
		"search_depth": searchDepth,
	}

	if maxResults := getInt(args, "max_results"); maxResults > 0 {
		payload["max_results"] = maxResults
	}
	if incDomains, ok := args["include_domains"].([]interface{}); ok {
		payload["include_domains"] = incDomains
	}
	if excDomains, ok := args["exclude_domains"].([]interface{}); ok {
		payload["exclude_domains"] = excDomains
	}

	jsonPayload, errMarshal := json.Marshal(payload)
	if errMarshal != nil {
		return err(fmt.Sprintf("Failed to marshal request payload: %v", errMarshal))
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, errReq := http.NewRequestWithContext(ctx, "POST", "https://api.tavily.com/search", bytes.NewBuffer(jsonPayload))
	if errReq != nil {
		return err(fmt.Sprintf("Failed to create request: %v", errReq))
	}
	req.Header.Set("Content-Type", "application/json")

	resp, errDo := client.Do(req)
	if errDo != nil {
		return err(fmt.Sprintf("Tavily API request failed: %v", errDo))
	}
	defer resp.Body.Close()

	body, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return err(fmt.Sprintf("Failed to read response body: %v", errRead))
	}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Tavily API error (HTTP %d): %s", resp.StatusCode, string(body)))
	}

	return ok(string(body))
}
