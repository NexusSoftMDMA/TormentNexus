package tools

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

// BraveSearchResponse represents the response from Brave Search API
type BraveSearchResponse struct {
	Web struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Snippet string `json:"description"`
		} `json:"results"`
	} `json:"web"`
}

// BraveImageResponse represents the response from Brave Image Search API
type BraveImageResponse struct {
	Results []struct {
		Title     string `json:"title"`
		URL       string `json:"url"`
		Source    string `json:"source"`
		Thumbnail string `json:"thumbnail"`
	} `json:"results"`
}

func getBraveAPIKey() string {
	key := os.Getenv("BRAVE_API_KEY")
	return key
}

func makeBraveRequest(ctx context.Context, apiURL string) ([]byte, error) {
	apiKey := getBraveAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("BRAVE_API_KEY environment variable not set")
}

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return nil, reqErr
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", apiKey)

	client := http.DefaultClient
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return nil, fetchErr
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Brave API returned status %d: %s", resp.StatusCode, string(body))
}

	return body, nil
}

// HandleBraveWebSearch performs a web search using the Brave Search API
func HandleBraveWebSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	count, _ :=getInt(args, "count")
	if count <= 0 {
		count = 10
	}
	if count > 20 {
		count = 20
	}

	offset, _ :=getInt(args, "offset")
	if offset < 0 {
		offset = 0
	}

	searchFreshness, _ :=getString(args, "search_freshness")

	params := url.Values{}
	params.Set("q", query)
	params.Set("count", fmt.Sprintf("%d", count))
	params.Set("offset", fmt.Sprintf("%d", offset))
	if searchFreshness != "" {
		params.Set("freshness", searchFreshness)

	apiURL := "https://api.search.brave.com/res/v1/web/search?" + params.Encode()

	body, apiErr := makeBraveRequest(ctx, apiURL)
	if apiErr != nil {
		return err(apiErr.Error())
}

	var searchResp BraveSearchResponse
	parseErr := json.Unmarshal(body, &searchResp)
	if parseErr != nil {
		return err(parseErr.Error())
}

	if len(searchResp.Web.Results) == 0 {
		return ok("No results found.")
}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Search results for: %s\n\n", query))

	for i, item := range searchResp.Web.Results {
		result.WriteString(fmt.Sprintf("%d. %s\n", i+1, item.Title))
		result.WriteString(fmt.Sprintf("   URL: %s\n", item.URL))
		result.WriteString(fmt.Sprintf("   %s\n\n", item.Snippet))

	return ok(result.String())
}

}
}

// HandleBraveImageSearch performs an image search using the Brave Search API
func HandleBraveImageSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	count, _ :=getInt(args, "count")
	if count <= 0 {
		count = 10
	}
	if count > 20 {
		count = 20
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("count", fmt.Sprintf("%d", count))

	apiURL := "https://api.search.brave.com/res/v1/images/search?" + params.Encode()

	body, apiErr := makeBraveRequest(ctx, apiURL)
	if apiErr != nil {
		return err(apiErr.Error())
}

	var imageResp BraveImageResponse
	parseErr := json.Unmarshal(body, &imageResp)
	if parseErr != nil {
		return err(parseErr.Error())
}

	if len(imageResp.Results) == 0 {
		return ok("No image results found.")
}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Image search results for: %s\n\n", query))

	for i, item := range imageResp.Results {
		result.WriteString(fmt.Sprintf("%d. %s\n", i+1, item.Title))
		result.WriteString(fmt.Sprintf("   Image URL: %s\n", item.URL))
		result.WriteString(fmt.Sprintf("   Source: %s\n", item.Source))
		if item.Thumbnail != "" {
			result.WriteString(fmt.Sprintf("   Thumbnail: %s\n", item.Thumbnail))

		result.WriteString("\n")

	return ok(result.String())
}

}
}

// HandleBraveNewsSearch performs a news search using the Brave Search API
func HandleBraveNewsSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	count, _ :=getInt(args, "count")
	if count <= 0 {
		count = 10
	}
User interrupted the output. Let me continue from where it was cut off and complete the remaining handlers.

	if count > 20 {
		count = 20
	}

	offset, _ :=getInt(args, "offset")
	if offset < 0 {
		offset = 0
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("count", fmt.Sprintf("%d", count))
	params.Set("offset", fmt.Sprintf("%d", offset))

	apiURL := "https://api.search.brave.com/res/v1/news/search?" + params.Encode()

	body, apiErr := makeBraveRequest(ctx, apiURL)
	if apiErr != nil {
		return err(apiErr.Error())
}

	var searchResp BraveSearchResponse
	parseErr := json.Unmarshal(body, &searchResp)
	if parseErr != nil {
		return err(parseErr.Error())
}

	if len(searchResp.Web.Results) == 0 {
		return ok("No news results found.")
}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("News results for: %s\n\n", query))

	for i, item := range searchResp.Web.Results {
		result.WriteString(fmt.Sprintf("%d. %s\n", i+1, item.Title))
		result.WriteString(fmt.Sprintf("   URL: %s\n", item.URL))
		result.WriteString(fmt.Sprintf("   %s\n\n", item.Snippet))

	return ok(result.String())
}

}

// HandleBraveSuggest provides search suggestions using the Brave Search API
func HandleBraveSuggest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	count, _ :=getInt(args, "count")
	if count <= 0 {
		count = 5
	}
	if count > 10 {
		count = 10
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("count", fmt.Sprintf("%d", count))

	apiURL := "https://api.search.brave.com/res/v1/suggest?" + params.Encode()

	body, apiErr := makeBraveRequest(ctx, apiURL)
	if apiErr != nil {
		return err(apiErr.Error())
}

	var suggestions struct {
		Results []struct {
			Query string `json:"query"`
		} `json:"results"`
	}
	parseErr := json.Unmarshal(body, &suggestions)
	if parseErr != nil {
		return err(parseErr.Error())
}

	if len(suggestions.Results) == 0 {
		return ok("No suggestions found.")
}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Search suggestions for: %s\n\n", query))

	for i, item := range suggestions vic.Results {
		result.WriteString(fmt.Sprintf("%d. %s\n", i+1, item.Query))

	return ok(result.String())
}

}

// HandleBraveSpellcheck provides spell check using the Brave Search API
func HandleBraveSpellcheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	params := url.Values{}
	params.Set("q", query)

	apiURL := "https://api.search.brave.com/res/v1/spellcheck?" + params.Encode()

	body, apiErr := makeBraveRequest(ctx, apiURL)
	if apiErr != nil {
		return err(apiErr.Error())
}

	var spellcheck struct {
		Query   string `json:"query"`
		Correct bool   `json:"correct"`
	}
	parseErr := json.Unmarshal(body, &spellcheck)
	if parseErr != nil {
		return err(parseErr.Error())
}

	if spellcheck.Correct {
		return ok(fmt.Sprintf("The query '%s' is spelled correctly.", query))
}

	return ok(fmt.Sprintf("The query '%s' may have spelling issues. Suggested: %s", query, spellcheck.Query))
}