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
)

func HandleTavilySearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	depth, _ :=getString(args, "search_depth")
	if depth == "" {
		depth = "basic"
	}
	includeAnswer, _ :=getBool(args, "include_answer")
	maxResults, _ :=getInt(args, "max_results")
	if maxResults < 1 {
		maxResults = 5
	}
	apiKey := os.Getenv("TAVILY_API_KEY")
	if apiKey == "" {
		return err("TAVILY_API_KEY not set")
}

	body := map[string]interface{}{
		"api_key":        apiKey,
		"query":          query,
		"search_depth":   depth,
		"include_answer": includeAnswer,
		"max_results":    maxResults,
	}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.tavily.com/search", bytes.NewReader(jsonBody))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(respBody)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(respBody, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	answer, found := result["answer"]
	if found {
		return success(fmt.Sprintf("Answer: %v", answer))
}

	results, found := result["results"]
	if !found {
		return err("no results in response")
}

	return success(fmt.Sprintf("Results: %v", results))
}
