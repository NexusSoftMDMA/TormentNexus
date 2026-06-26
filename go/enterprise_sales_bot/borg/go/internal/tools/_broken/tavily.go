package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TavilySearchArgs holds the arguments for the search tool
type TavilySearchArgs struct {
	Query   string `json:"query"`
	MaxResults int `json:"max_results,omitempty"`
	SearchDepth string `json:"search_depth,omitempty"`
	IncludeAnswer bool `json:"include_answer,omitempty"`
	IncludeImages bool `json:"include_images,omitempty"`
	IncludeRawContent bool `json:"include_raw_content,omitempty"`
}

// TavilySearchResult represents the response from the Tavily API
type TavilySearchResult struct {
	Query      string `json:"query"`
	Answer     string `json:"answer,omitempty"`
	Images     []string `json:"images,omitempty"`
	Results    []SearchResultItem `json:"results"`
	RawContent []string `json:"raw_content,omitempty"`
}

// SearchResultItem represents a single search result
type SearchResultItem struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Content     string `json:"content"`
	Score       float64 `json:"score"`
	PublishedDate string `json:"published_date,omitempty"`
}

// Client for Tavily API
var tavilyClient = http.DefaultClient

// HandleTavilySearch implements the tavily search tool
func HandleTavilySearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	maxResults, _ :=getInt(args, "max_results")
	if maxResults == 0 {
		maxResults = 5
	}

	searchDepth, _ :=getString(args, "search_depth")
	if searchDepth == "" {
		searchDepth = "basic"
	}

	includeAnswer, _ :=getBool(args, "include_answer")
	includeImages, _ :=getBool(args, "include_images")
	includeRawContent, _ :=getBool(args, "include_raw_content")

	// Build request payload
	payload := map[string]interface{}{
		"query":           query,
		"api_key":         apiKey,
		"max_results":     maxResults,
		"search_depth":    searchDepth,
		"include_answer":  includeAnswer,
		"include_images":  includeImages,
		"include_raw_content": includeRawContent,
	}

	jsonPayload, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal request: %v", e))
}

	// Create HTTP request
	req, reqErr := http.NewRequestWithContext(ctx, "POST", "https://api.tavily.com/search", strings.NewReader(string(jsonPayload)))
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, fetchErr := tavilyClient.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to execute request: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	// Parse response
	var result TavilySearchResult
	parseErr := json.NewDecoder(resp.Body).Decode(&result)
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	// Format output
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Search Results for: %s\n\n", query))

	if result.Answer != "" {
		output.WriteString(fmt.Sprintf("Answer: %s\n\n", result.Answer))

	output.WriteString("Results:\n")
	for i, item := range result.Results {
		output.WriteString(fmt.Sprintf("%d. %s\n   URL: %s\n   Content: %s\n\n", i+1, item.Title, item.URL, item.Content))

	if len(result.Images) > 0 {
		output.WriteString("Images:\n")
		for i, img := range result.Images {
			output.WriteString(fmt.Sprintf("%d. %s\n", i+1, img))

		output.WriteString("\n")

	return ok(output.String())
}

}
}
}
}

// HandleTavilySearchWithUrl implements a simplified search that just returns URLs
func HandleTavilySearchWithUrl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	// Build request payload
	payload := map[string]interface{}{
		"query":       query,
		"api_key":     apiKey,
		"max_results": 3,
		"search_depth": "basic",
	}

	jsonPayload, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal request: %v", e))
}

	// Create HTTP request
	req, reqErr := http.NewRequestWithContext(ctx, "POST", "https://api.tavily.com/search", strings.NewReader(string(jsonPayload)))
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, fetchErr := tavilyClient.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to execute request: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	// Parse response
	var result TavilySearchResult
	parseErr := json.NewDecoder(resp.Body).Decode(&result)
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	// Format output - just URLs
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Top URLs for: %s\n", query))
	for i, item := range result.Results {
		output.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, item.Title, item.URL))

	return ok(output.String())
}

}

// HandleTavilyExtract implements the extract tool to get content from specific URLs
func HandleTavilyExtract(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	urls, _ :=getString(args, "urls")
	if urls == "" {
		return err("urls is required")
}

	// Parse comma-separated URLs
	urlList := strings.Split(urls, ",")
	var cleanUrls []string
	for _, u := range urlList {
		cleanUrls = append(cleanUrls, strings.TrimSpace(u))

	// Build request payload
	payload := map[string]interface{}{
		"api_key": apiKey,
		"urls":    cleanUrls,
	}

	jsonPayload, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal request: %v", e))
}

	// Create HTTP request
	req, reqErr := http.NewRequestWithContext(ctx, "POST", "https://api.tavily.com/extract", strings.NewReader(string(jsonPayload)))
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, fetchErr := tavilyClient.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to execute request: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	// Parse response
	var result struct {
		Results []struct {
			URL     string `json:"url"`
			Content string `json:"content"`
			Title   string `json:"title"`
		} `json:"results"`
	}
	parseErr := json.NewDecoder(resp.Body).Decode(&result)
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	// Format output
	var output strings.Builder
	output.WriteString("Extracted Content:\n\n")
	for i, item := range result.Results {
		output.WriteString(fmt.Sprintf("%d. %s\n   URL: %s\n   Content: %s\n\n", i+1, item.Title, item.URL, item.Content))

	return ok(output.String())
}

}
}

// HandleTavilyNews implements the news search tool
func HandleTavilyNews(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	// Build request payload
	payload := map[string]interface{}{
		"query":       query,
		"api_key":     apiKey,
		"search_type": "news",
		"max_results": 5,
	}

	jsonPayload, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal request: %v", e))
}

	// Create HTTP request
	req, reqErr := http.NewRequestWithContext(ctx, "POST", "https://api.tavily.com/search", strings.NewReader(string(jsonPayload)))
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, fetchErr := tavilyClient.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to execute request: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	// Parse response
	var result TavilySearchResult
	parseErr := json.NewDecoder(resp.Body).Decode(&result)
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	// Format output
	var output strings.Builder
	output.WriteString(fmt.Sprintf("News Results for: %s\n\n", query))
	for i, item := range result.Results {
		output.WriteString(fmt.Sprintf("%d. %s\n   URL: %s\n   Content: %s\n", i+1, item.Title, item.URL, item.Content))
		if item.PublishedDate != "" {
			output.WriteString(fmt.Sprintf("   Date: %s\n", item.PublishedDate))

		output.WriteString("\n")

	return ok(output.String())
}
}
}