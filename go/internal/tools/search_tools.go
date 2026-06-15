//go:build ignore
// +build ignore

package tools

/**
 * @file search_tools.go
 * @module go/internal/tools
 *
 * WHAT: DuckDuckGo search tools for parity with Claude Code / Codex.
 * Provides HandleDDGSearch (web search via DDG Instant Answer API)
 * and HandleDDGFetchContent (fetch content from a URL).
 *
 * WHY: Parity — Claude Code and Codex expose search/fetch_content tools.
 */

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

// ddgInstantAnswer is the response from the DuckDuckGo Instant Answer API.
type ddgInstantAnswer struct {
	Abstract       string          `json:"Abstract"`
	AbstractSource string          `json:"AbstractSource"`
	AbstractURL    string          `json:"AbstractURL"`
	Answer         string          `json:"Answer"`
	AnswerType     string          `json:"AnswerType"`
	Definition     string          `json:"Definition"`
	DefinitionSource string        `json:"DefinitionSource"`
	Heading        string          `json:"Heading"`
	Image          string          `json:"Image"`
	Redirect       string          `json:"Redirect"`
	Results        []ddgResult     `json:"Results"`
	RelatedTopics  []ddgRelated    `json:"RelatedTopics"`
	Type           string          `json:"Type"`
	SourceDomain   string          `json:"SourceDomain"`
}

type ddgResult struct {
	FirstURL string `json:"FirstURL"`
	Icon     struct {
		Height string `json:"Height"`
		URL    string `json:"URL"`
		Width  string `json:"Width"`
	} `json:"Icon"`
	Result  string `json:"Result"`
	Text    string `json:"Text"`
	URL     string `json:"URL"`
}

type ddgRelated struct {
	Text     string       `json:"Text"`
	FirstURL string       `json:"FirstURL"`
	Result   string       `json:"Result"`
	Topics   []ddgRelated `json:"Topics,omitempty"`
	Name     string       `json:"Name,omitempty"`
}

// HandleDDGSearch performs a web search using DuckDuckGo's Instant Answer API.
// No API key required. Returns a formatted summary of results.
func HandleDDGSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q", "search")
	if query == "" {
		return err("query is required")
	}

	maxResults := getInt(args, "max_results", "maxResults", "limit")
	if maxResults <= 0 || maxResults > 20 {
		maxResults = 5
	}

	timeoutSec := getInt(args, "timeout")
	if timeoutSec <= 0 {
		timeoutSec = 15
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	apiURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1",
		url.QueryEscape(query))

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("Error creating request: %v", reqErr))
	}
	req.Header.Set("User-Agent", "TormentNexus/Search-Parity/1.0")

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("Search request failed: %v", fetchErr))
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("Error reading response: %v", readErr))
	}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Search API returned HTTP %d: %s", resp.StatusCode, string(body)))
	}

	var result ddgInstantAnswer
	if parseErr := json.Unmarshal(body, &result); parseErr != nil {
		return err(fmt.Sprintf("Error parsing search results: %v", parseErr))
	}

	var sb strings.Builder

	// Abstract / Answer
	if result.Abstract != "" {
		sb.WriteString(fmt.Sprintf("## %s\n\n%s\n\n**Source:** %s\n\n",
			result.Heading, result.Abstract, result.AbstractSource))
	}

	if result.Answer != "" {
		sb.WriteString(fmt.Sprintf("**Answer:** %s\n\n", result.Answer))
	}

	if result.Definition != "" {
		sb.WriteString(fmt.Sprintf("**Definition:** %s (%s)\n\n",
			result.Definition, result.DefinitionSource))
	}

	// Results
	count := 0
	for _, r := range result.Results {
		if count >= maxResults {
			break
		}
		url := r.FirstURL
		if url == "" {
			url = r.URL
		}
		text := stripHTML(r.Text)
		if text == "" {
			text = stripHTML(r.Result)
		}
		if text != "" {
			sb.WriteString(fmt.Sprintf("%d. [%s](%s)\n   %s\n\n", count+1, url, url, text))
			count++
		}
	}

	// Related Topics
	for _, rt := range result.RelatedTopics {
		if count >= maxResults {
			break
		}
		if rt.Text != "" {
			url := rt.FirstURL
			text := stripHTML(rt.Text)
			sb.WriteString(fmt.Sprintf("%d. [%s](%s)\n   %s\n\n", count+1, url, url, text))
			count++
		} else if len(rt.Topics) > 0 {
			// Topic category heading
			for _, sub := range rt.Topics {
				if count >= maxResults {
					break
				}
				text := stripHTML(sub.Text)
				if text != "" {
					sb.WriteString(fmt.Sprintf("%d. [%s](%s)\n   %s\n\n", count+1, sub.FirstURL, sub.FirstURL, text))
					count++
				}
			}
		}
	}

	if sb.Len() == 0 {
		return ok(fmt.Sprintf("No results found for: %s", query))
	}

	return ok(strings.TrimSpace(sb.String()))
}

// HandleDDGFetchContent fetches the content of a URL and returns it as text.
// This is a parity tool matching the "fetch_content" tool in other harnesses.
func HandleDDGFetchContent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ := getString(args, "url", "uri", "target")
	if urlStr == "" {
		return err("url is required")
	}

	timeoutSec := getInt(args, "timeout")
	if timeoutSec <= 0 {
		timeoutSec = 30
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	req, reqErr := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("Error creating request: %v", reqErr))
	}
	req.Header.Set("User-Agent", "TormentNexus/FetchContent-Parity/1.0")

	resp, doErr := http.DefaultClient.Do(req)
	if doErr != nil {
		return err(fmt.Sprintf("Fetch request failed: %v", doErr))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status))
	}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("Error reading response: %v", readErr))
	}

	contentType := resp.Header.Get("Content-Type")

	// Only return text content
	if !strings.HasPrefix(contentType, "text/") &&
		!strings.HasPrefix(contentType, "application/json") &&
		!strings.HasPrefix(contentType, "application/xml") &&
		!strings.HasPrefix(contentType, "application/xhtml+xml") {
		return ok(fmt.Sprintf("Content-Type is %s (%d bytes, not displayed as text)", contentType, len(body)))
	}

	content := string(body)
	if len(content) > 100000 {
		content = content[:100000] + "\n...[Content truncated at 100KB]"
	}

	return ok(content)
}

// stripHTML removes HTML tags from a string.
func stripHTML(s string) string {
	var sb strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			sb.WriteRune(r)
		}
	}
	return strings.TrimSpace(sb.String())
}
