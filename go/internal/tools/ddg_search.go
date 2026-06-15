//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// SearchResult matches the python SearchResult dataclass
type SearchResult struct {
	Title    string
	Link     string
	Snippet  string
	Position int
}

// HandleDDGSearch performs a web search using DuckDuckGo's HTML endpoint.
func HandleDDGSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q")
	if query == "" {
		return err("query parameter is required")
	}

	maxResults := getInt(args, "max_results", "maxResults", "limit")
	if maxResults <= 0 {
		maxResults = 10
	} else if maxResults > 20 {
		maxResults = 20
	}

	region, _ := getString(args, "region")

	// Prepare POST data
	formData := url.Values{}
	formData.Set("q", query)
	formData.Set("b", "")
	formData.Set("kl", region)
	formData.Set("kp", "-1") // moderate safe search

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, errReq := http.NewRequestWithContext(ctx, "POST", "https://html.duckduckgo.com/html", strings.NewReader(formData.Encode()))
	if errReq != nil {
		return err(fmt.Sprintf("Failed to create request: %v", errReq))
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, errDo := client.Do(req)
	if errDo != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", errDo))
	}
	defer resp.Body.Close()

	body, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return err(fmt.Sprintf("Failed to read response body: %v", errRead))
	}

	htmlStr := string(body)

	// Robust regex-based HTML parsing for DuckDuckGo HTML results
	resultRegex := regexp.MustCompile(`(?s)<div class="web-result[^"]*result.*?">(.*?)</div>\s*</div>`)
	titleRegex := regexp.MustCompile(`(?s)<a class="result__url"[^>]*href="([^"]*)"[^>]*>(.*?)</a>`)
	snippetRegex := regexp.MustCompile(`(?s)<a class="result__snippet"[^>]*>(.*?)</a>`)

	matches := resultRegex.FindAllStringSubmatch(htmlStr, -1)
	if len(matches) == 0 {
		resultRegex = regexp.MustCompile(`(?s)<div class="result.*?">(.*?)</div>\s*</div>`)
		matches = resultRegex.FindAllStringSubmatch(htmlStr, -1)
	}

	var results []SearchResult
	for _, m := range matches {
		innerHtml := m[1]

		titleMatches := titleRegex.FindStringSubmatch(innerHtml)
		if len(titleMatches) < 3 {
			continue
		}

		link := titleMatches[1]
		title := cleanHTMLTags(titleMatches[2])

		if strings.Contains(link, "uddg=") {
			parts := strings.Split(link, "uddg=")
			if len(parts) > 1 {
				subParts := strings.Split(parts[1], "&")
				if decoded, errDec := url.QueryUnescape(subParts[0]); errDec == nil {
					link = decoded
				}
			}
		}

		snippet := ""
		snippetMatches := snippetRegex.FindStringSubmatch(innerHtml)
		if len(snippetMatches) > 1 {
			snippet = cleanHTMLTags(snippetMatches[1])
		}

		results = append(results, SearchResult{
			Title:    strings.TrimSpace(title),
			Link:     strings.TrimSpace(link),
			Snippet:  strings.TrimSpace(snippet),
			Position: len(results) + 1,
		})

		if len(results) >= maxResults {
			break
		}
	}

	if len(results) == 0 {
		return ok("No results were found for your search query. Please try rephrasing your search.")
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Found %d search results:\n\n", len(results)))
	for _, r := range results {
		builder.WriteString(fmt.Sprintf("%d. %s\n", r.Position, r.Title))
		builder.WriteString(fmt.Sprintf("   URL: %s\n", r.Link))
		builder.WriteString(fmt.Sprintf("   Summary: %s\n\n", r.Snippet))
	}

	return ok(builder.String())
}

// HandleDDGFetchContent fetches content from a webpage, strips formatting, and paginates.
func HandleDDGFetchContent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ := getString(args, "url", "uri")
	if urlStr == "" {
		return err("url parameter is required")
	}

	startIndex := getInt(args, "start_index", "startIndex", "offset")
	maxLength := getInt(args, "max_length", "maxLength", "limit")
	if maxLength <= 0 {
		maxLength = 8000
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, errReq := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if errReq != nil {
		return err(fmt.Sprintf("Failed to create request: %v", errReq))
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, errDo := client.Do(req)
	if errDo != nil {
		return err(fmt.Sprintf("Fetch failed: %v", errDo))
	}
	defer resp.Body.Close()

	body, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return err(fmt.Sprintf("Failed to read response: %v", errRead))
	}

	rawText := string(body)

	cleanRegexes := []*regexp.Regexp{
		regexp.MustCompile(`(?s)<script.*?>.*?</script>`),
		regexp.MustCompile(`(?s)<style.*?>.*?</style>`),
		regexp.MustCompile(`(?s)<nav.*?>.*?</nav>`),
		regexp.MustCompile(`(?s)<header.*?>.*?</header>`),
		regexp.MustCompile(`(?s)<footer.*?>.*?</footer>`),
	}

	for _, re := range cleanRegexes {
		rawText = re.ReplaceAllString(rawText, " ")
	}

	cleanedText := cleanHTMLTags(rawText)

	spaceRegex := regexp.MustCompile(`\s+`)
	cleanedText = spaceRegex.ReplaceAllString(cleanedText, " ")
	cleanedText = strings.TrimSpace(cleanedText)

	totalLen := len(cleanedText)

	if startIndex < 0 {
		startIndex = 0
	}
	if startIndex >= totalLen {
		return ok(fmt.Sprintf("\n\n---\n[Content info: Showing characters %d-%d of %d total]", totalLen, totalLen, totalLen))
	}

	endIndex := startIndex + maxLength
	if endIndex > totalLen {
		endIndex = totalLen
	}

	paginatedText := cleanedText[startIndex:endIndex]
	isTruncated := endIndex < totalLen

	var metadata string
	if isTruncated {
		metadata = fmt.Sprintf("\n\n---\n[Content info: Showing characters %d-%d of %d total. Use start_index=%d to see more]", startIndex, endIndex, totalLen, endIndex)
	} else {
		metadata = fmt.Sprintf("\n\n---\n[Content info: Showing characters %d-%d of %d total]", startIndex, endIndex, totalLen)
	}

	return ok(paginatedText + metadata)
}

// Helper to remove HTML tags using regex
func cleanHTMLTags(src string) string {
	tagRegex := regexp.MustCompile(`(?s:<[^>]*>)`)
	src = tagRegex.ReplaceAllString(src, " ")
	src = strings.ReplaceAll(src, "&nbsp;", " ")
	src = strings.ReplaceAll(src, "&amp;", "&")
	src = strings.ReplaceAll(src, "&lt;", "<")
	src = strings.ReplaceAll(src, "&gt;", ">")
	src = strings.ReplaceAll(src, "&quot;", "\"")
	return src
}
