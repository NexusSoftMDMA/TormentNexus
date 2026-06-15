//go:build ignore
// +build ignore

package tools

/**
 * @file exa.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Exa search (https://exa.ai).
 * Replaces the SSE-based `exa` MCP server entry in mcp.json.
 *
 * Improvements over original:
 *  - No external process/SSE dependency.
 *  - Unified tool interface (exa_search, exa_find_similar, exa_get_contents).
 *  - Full context support & timeout control.
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

const exaBaseURL = "https://api.exa.ai"

func exaAPIKey() string {
	return os.Getenv("EXA_API_KEY")
}

func exaPost(ctx context.Context, endpoint string, payload interface{}) (map[string]interface{}, error) {
	apiKey := exaAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("EXA_API_KEY environment variable is not set")
	}

	data, _ := json.Marshal(payload)
	req, e := http.NewRequestWithContext(ctx, "POST", exaBaseURL+endpoint, bytes.NewBuffer(data))
	if e != nil {
		return nil, e
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Do(req)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Exa API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return nil, fmt.Errorf("failed to parse Exa response: %v", e)
	}
	return result, nil
}

// HandleExaSearch performs a neural/keyword web search via Exa API.
// Tool: exa_search
func HandleExaSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q")
	if query == "" {
		return err("query parameter is required")
	}

	payload := map[string]interface{}{
		"query": query,
	}

	if numResults := getInt(args, "numResults", "num_results", "count"); numResults > 0 {
		payload["numResults"] = numResults
	} else {
		payload["numResults"] = 10
	}

	if searchType, ok := args["type"].(string); ok && searchType != "" {
		payload["type"] = searchType // "neural" or "keyword"
	}

	if useAutoPrompt, ok := args["useAutoprompt"].(bool); ok {
		payload["useAutoprompt"] = useAutoPrompt
	}

	if includeDomains, ok := args["includeDomains"].([]interface{}); ok {
		payload["includeDomains"] = includeDomains
	}

	if excludeDomains, ok := args["excludeDomains"].([]interface{}); ok {
		payload["excludeDomains"] = excludeDomains
	}

	if startDate, _ := getString(args, "startPublishedDate"); startDate != "" {
		payload["startPublishedDate"] = startDate
	}

	if endDate, _ := getString(args, "endPublishedDate"); endDate != "" {
		payload["endPublishedDate"] = endDate
	}

	// Include contents if requested
	if includeText, ok := args["contents"].(map[string]interface{}); ok {
		payload["contents"] = includeText
	}

	result, e := exaPost(ctx, "/search", payload)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleExaFindSimilar finds pages similar to a given URL.
// Tool: exa_find_similar
func HandleExaFindSimilar(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ := getString(args, "url", "uri")
	if urlStr == "" {
		return err("url parameter is required")
	}

	payload := map[string]interface{}{
		"url": urlStr,
	}

	if numResults := getInt(args, "numResults", "num_results", "count"); numResults > 0 {
		payload["numResults"] = numResults
	} else {
		payload["numResults"] = 10
	}

	if includeDomains, ok := args["includeDomains"].([]interface{}); ok {
		payload["includeDomains"] = includeDomains
	}

	if excludeDomains, ok := args["excludeDomains"].([]interface{}); ok {
		payload["excludeDomains"] = excludeDomains
	}

	if excludeSourceDomain, ok := args["excludeSourceDomain"].(bool); ok {
		payload["excludeSourceDomain"] = excludeSourceDomain
	}

	result, e := exaPost(ctx, "/findSimilar", payload)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleExaGetContents retrieves full page contents for a list of Exa result IDs.
// Tool: exa_get_contents
func HandleExaGetContents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ids, found := args["ids"].([]interface{})
	if !found || len(ids) == 0 {
		// Also accept single url
		urlStr, _ := getString(args, "url", "uri")
		if urlStr == "" {
			return err("ids (array) or url parameter is required")
		}
		ids = []interface{}{urlStr}
	}

	payload := map[string]interface{}{
		"ids": ids,
	}

	if text, ok := args["text"].(map[string]interface{}); ok {
		payload["text"] = text
	} else {
		payload["text"] = map[string]interface{}{"maxCharacters": 5000}
	}

	if highlights, ok := args["highlights"].(map[string]interface{}); ok {
		payload["highlights"] = highlights
	}

	if summary, ok := args["summary"].(map[string]interface{}); ok {
		payload["summary"] = summary
	}

	result, e := exaPost(ctx, "/contents", payload)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

