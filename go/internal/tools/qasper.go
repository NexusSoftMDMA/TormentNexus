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

var (
	// Compile regex patterns once at package init
	arxivIDPattern = regexp.MustCompile(`^\d{4}\.\d{4,5}$`)
)

func HandleGetPaperInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	paperID, _ :=getString(args, "paper_id")
	if paperID == "" {
		return err("paper_id is required")
}

	// Validate arxiv ID format
	if !arxivIDPattern.MatchString(paperID) {
		return err("invalid paper_id format. Expected format: 1234.5678")
}

	// Construct arxiv API URL
	apiURL := fmt.Sprintf("https://export.arxiv.org/api/query?id_list=%s", paperID)

	client := http.Client{Timeout: 30 * time.Second}
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch paper info: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("arxiv API returned status: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response body: %v", readErr))
}

	// Parse the XML response (simplified parsing for demonstration)
	// In a real implementation, you'd use proper XML parsing
	content := string(body)
	title := extractField(content, "<title>", "</title>")
	authors := extractField(content, "<author>", "</author>")
	summary := extractField(content, "<summary>", "</summary>")

	if title == "" {
		return err("failed to extract paper information from response")
}

	result := fmt.Sprintf("Title: %s\nAuthors: %s\nSummary: %s",
		strings.TrimSpace(title),
		strings.TrimSpace(authors),
		strings.TrimSpace(summary))

	return ok(result)
}

func HandleSearchPapers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	maxResults, _ :=getInt(args, "max_results")
	if maxResults == 0 {
		maxResults = 5
	}

	if query == "" {
		return err("query is required")
}

	// URL encode the query
	encodedQuery := url.QueryEscape(query)
	apiURL := fmt.Sprintf("https://export.arxiv.org/api/query?search_query=%s&start=0&max_results=%d",
		encodedQuery, maxResults)

	client := http.Client{Timeout: 30 * time.Second}
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to search papers: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("arxiv API returned status: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response body: %v", readErr))
}

	// Parse the XML response (simplified parsing for demonstration)
	content := string(body)
	entries := strings.Split(content, "<entry>")
	if len(entries) <= 1 {
		return ok("No papers found matching your query")
}

	var results strings.Builder
	for i, entry := range entries[1:] { // Skip first empty entry
		if i >= maxResults {
			break
		}

		title := extractField(entry, "<title>", "</title>")
		id := extractField(entry, "<id>http://arxiv.org/abs/", "</id>")
		summary := extractField(entry, "<summary>", "</summary>")

		if title != "" && id != "" {
			results.WriteString(fmt.Sprintf("Paper %d:\nID: %s\nTitle: %s\nSummary: %s\n\n",
				i+1, id, strings.TrimSpace(title), strings.TrimSpace(summary)))

	}

	if results.Len() == 0 {
		return ok("No papers found matching your query")
}

	return ok(results.String())
}

}

func HandleGetPaperPDF(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	paperID, _ :=getString(args, "paper_id")
	if paperID == "" {
		return err("paper_id is required")
}

	// Validate arxiv ID format
	if !arxivIDPattern.MatchString(paperID) {
		return err("invalid paper_id format. Expected format: 1234.5678")
}

	// Construct PDF URL
	pdfURL := fmt.Sprintf("https://arxiv.org/pdf/%s.pdf", paperID)

	client := http.Client{Timeout: 30 * time.Second}
	req, reqErr := http.NewRequestWithContext(ctx, "HEAD", pdfURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to check PDF availability: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("PDF not available for paper %s", paperID))
}

	return ok(pdfURL)
}

// Helper function to extract content between XML tags
func extractField(content, startTag, endTag string) string {
	start := strings.Index(content, startTag)
	if start == -1 {
		return ""
	}
	start += len(startTag)

	end := strings.Index(content[start:], endTag)
	if end == -1 {
		return ""
	}

	return content[start : start+end]
}