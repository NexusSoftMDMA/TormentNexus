//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// HandleFetch implements the fetch tool functionality natively.
// It retrieves the web page at a URL, strips unwanted tags, converts to markdown/text, and supports pagination.
func HandleFetch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ := getString(args, "url", "uri")
	if urlStr == "" {
		return err("url parameter is required")
	}

	headers, _ := args["headers"].(map[string]interface{})
	maxLength := getInt(args, "max_length", "maxLength")
	if maxLength <= 0 {
		maxLength = 5000
	}

	startIndex := getInt(args, "start_index", "startIndex")
	rawVal := getBool(args, "raw")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, errReq := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if errReq != nil {
		return err(fmt.Sprintf("Failed to create request: %v", errReq))
	}

	// Add custom headers if provided
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	for k, v := range headers {
		if valStr, ok := v.(string); ok {
			req.Header.Set(k, valStr)
		}
	}

	resp, errDo := client.Do(req)
	if errDo != nil {
		return err(fmt.Sprintf("Fetch failed: %v", errDo))
	}
	defer resp.Body.Close()

	body, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return err(fmt.Sprintf("Failed to read response: %v", errRead))
	}

	content := string(body)

	if !rawVal {
		// Clean up the HTML to approximate markdown/readable plain text
		cleanRegexes := []*regexp.Regexp{
			regexp.MustCompile(`(?s)<script.*?>.*?</script>`),
			regexp.MustCompile(`(?s)<style.*?>.*?</style>`),
			regexp.MustCompile(`(?s)<nav.*?>.*?</nav>`),
			regexp.MustCompile(`(?s)<header.*?>.*?</header>`),
			regexp.MustCompile(`(?s)<footer.*?>.*?</footer>`),
		}
		for _, re := range cleanRegexes {
			content = re.ReplaceAllString(content, " ")
		}

		content = cleanHTMLTags(content)

		// Replace multiple spaces/newlines with single ones
		spaceRegex := regexp.MustCompile(`\s+`)
		content = spaceRegex.ReplaceAllString(content, " ")
		content = strings.TrimSpace(content)
	}

	totalLen := len(content)
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

	paginatedText := content[startIndex:endIndex]
	isTruncated := endIndex < totalLen

	var metadata string
	if isTruncated {
		metadata = fmt.Sprintf("\n\n---\n[Content info: Showing characters %d-%d of %d total. Use start_index=%d to see more]", startIndex, endIndex, totalLen, endIndex)
	} else {
		metadata = fmt.Sprintf("\n\n---\n[Content info: Showing characters %d-%d of %d total]", startIndex, endIndex, totalLen)
	}

	return ok(paginatedText + metadata)
}
