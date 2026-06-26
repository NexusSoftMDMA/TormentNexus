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

// HandleSearchPapers searches for academic papers using the Semantic Scholar API.
func HandleSearchPapers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	client := http.DefaultClient
	
	// Construct API URL
	apiURL := fmt.Sprintf("https://api.semanticscholar.org/graph/v1/paper/search?query=%s&fields=title,abstract,url,year", url.QueryEscape(query))

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API request failed with status code: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	var searchResult struct {
		Data []struct {
			Title    string `json:"title"`
			Abstract string `json:"abstract"`
			Url      string `json:"url"`
			Year     int    `json:"year"`
		} `json:"data"`
	}

	if jsonErr := json.Unmarshal(body, &searchResult); jsonErr != nil {
		return err(jsonErr.Error())
}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Found %d papers for query '%s':\n\n", len(searchResult.Data), query))

	for i, paper := range searchResult.Data {
		builder.WriteString(fmt.Sprintf("%d. **%s** (%d)\n", i+1, paper.Title, paper.Year))
		if paper.Url != "" {
			builder.WriteString(fmt.Sprintf("   URL: %s\n", paper.Url))

		if paper.Abstract != "" {
			// Truncate abstract if too long for display
			abs := paper.Abstract
			if len(abs) > 200 {
				abs = abs[:200] + "..."
			}
			builder.WriteString(fmt.Sprintf("   Abstract: %s\n", abs))

		builder.WriteString("\n")

	return ok(builder.String())
}

}
}
}

// HandleFetchUrl retrieves the raw content of a specified URL.
func HandleFetchUrl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetUrl, _ :=getString(args, "url")
	if targetUrl == "" {
		return err("url is required")
}

	client := http.DefaultClient

	req, reqErr := http.NewRequestWithContext(ctx, "GET", targetUrl, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("HTTP request failed with status code: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	return ok(string(body))
}

// HandleSummarize provides a basic extractive summary of the provided text.
func HandleSummarize(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	// Simple extractive summary: take the first 3 sentences.
	sentences := strings.Split(text, ". ")
	count := 3
	if len(sentences) < count {
		count = len(sentences)

	var summaryParts []string
	for i := 0; i < count; i++ {
		trimmed := strings.TrimSpace(sentences[i])
		if trimmed != "" {
			summaryParts = append(summaryParts, trimmed)

	}

	summary := strings.Join(summaryParts, ". ")
	if !strings.HasSuffix(summary, ".") {
		summary += "."
	}

	return ok(summary)
}
}
}