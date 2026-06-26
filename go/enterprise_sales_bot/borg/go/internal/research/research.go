package research

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ResearchService performs deep research on topics by searching the web
// and reading content from found URLs.
type ResearchService struct {
	httpClient *http.Client
	visited    map[string]bool
	mu         sync.Mutex
	mcpToolURL string // URL to call MCP tools for content fetching
}

// NewResearchService creates a new research service.
func NewResearchService(mcpToolURL string) *ResearchService {
	if mcpToolURL == "" {
		mcpToolURL = "http://localhost:7778/api/mcp/tools/call"
	}
	return &ResearchService{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		visited:    make(map[string]bool),
		mcpToolURL: mcpToolURL,
	}
}

// ResearchResult holds the result of a research operation.
type ResearchResult struct {
	Report      string   `json:"report"`
	SourcesUsed int      `json:"sourcesUsed"`
	URLsVisited []string `json:"urlsVisited"`
	Duration    string   `json:"duration"`
}

// Conduct performs research on a topic with the given depth.
func (rs *ResearchService) Conduct(ctx context.Context, topic string, depth int) (*ResearchResult, error) {
	start := time.Now()
	rs.mu.Lock()
	rs.visited = make(map[string]bool)
	rs.mu.Unlock()

	if depth <= 0 || depth > 10 {
		depth = 3
	}

	var reportParts []string
	reportParts = append(reportParts, fmt.Sprintf("# Research Report: %s", topic))

	// Search for sources
	urls, err := rs.search(ctx, topic, depth)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	reportParts = append(reportParts, fmt.Sprintf("Found %d primary sources.", len(urls)))

	// Read each URL
	var visited []string
	for i, target := range urls {
		rs.mu.Lock()
		if rs.visited[target] {
			rs.mu.Unlock()
			continue
		}
		rs.visited[target] = true
		rs.mu.Unlock()

		visited = append(visited, target)

		content, err := rs.fetchContent(ctx, target)
		if err != nil {
			reportParts = append(reportParts, fmt.Sprintf("\n## Source %d: %s\n[Error reading: %v]", i+1, target, err))
			continue
		}

		reportParts = append(reportParts, fmt.Sprintf("\n## Source %d: %s\n%s", i+1, target, truncate(content, 2000)))
	}

	elapsed := time.Since(start).Round(time.Second).String()

	return &ResearchResult{
		Report:      strings.Join(reportParts, "\n"),
		SourcesUsed: len(urls),
		URLsVisited: visited,
		Duration:    elapsed,
	}, nil
}

// Ingest reads a URL and stores its content for later retrieval.
func (rs *ResearchService) Ingest(ctx context.Context, rawURL string) (string, error) {
	content, err := rs.fetchContent(ctx, rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch %s: %w", rawURL, err)
	}

	summary := fmt.Sprintf("Ingested %s: %.200s", rawURL, content)
	return summary, nil
}

// search performs a web search using DuckDuckGo.
func (rs *ResearchService) search(ctx context.Context, query string, count int) ([]string, error) {
	// Use DuckDuckGo lite (no API key needed)
	searchURL := fmt.Sprintf("https://lite.duckduckgo.com/lite/?q=%s", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	resp, err := rs.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse DDG lite results (HTML with links)
	return parseDDGResults(string(body), count), nil
}

// parseDDGResults extracts URLs from DuckDuckGo lite HTML results.
func parseDDGResults(html string, max int) []string {
	var urls []string
	seen := make(map[string]bool)
	lines := strings.Split(html, "\n")

	for _, line := range lines {
		if strings.Contains(line, "<a ") && strings.Contains(line, "href=\"") {
			parts := strings.Split(line, "href=\"")
			if len(parts) < 2 {
				continue
			}
			href := strings.Split(parts[1], "\"")[0]
			if !strings.HasPrefix(href, "http") {
				continue
			}
			if seen[href] {
				continue
			}
			seen[href] = true
			urls = append(urls, href)
			if len(urls) >= max {
				break
			}
		}
	}
	return urls
}

// fetchContent retrieves and extracts text content from a URL.
func (rs *ResearchService) fetchContent(ctx context.Context, targetURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := rs.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Simple content extraction — strip HTML tags
	text := stripHTMLTags(string(body))
	return strings.TrimSpace(text), nil
}

// stripHTMLTags removes HTML tags from a string.
func stripHTMLTags(html string) string {
	var result strings.Builder
	inTag := false
	for _, ch := range html {
		if ch == '<' {
			inTag = true
			continue
		}
		if ch == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(ch)
		}
	}
	return result.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// --- JSON helpers for handler integration ---

type ResearchRequest struct {
	Topic string `json:"topic"`
	Depth int    `json:"depth"`
}

type IngestRequest struct {
	URL string `json:"url"`
}

func (rs *ResearchService) HandleConduct(w http.ResponseWriter, r *http.Request) {
	var req ResearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}
	if req.Topic == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing topic"})
		return
	}

	result, err := rs.Conduct(r.Context(), req.Topic, req.Depth)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result})
}

func (rs *ResearchService) HandleIngest(w http.ResponseWriter, r *http.Request) {
	var req IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}
	if req.URL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing url"})
		return
	}

	result, err := rs.Ingest(r.Context(), req.URL)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
