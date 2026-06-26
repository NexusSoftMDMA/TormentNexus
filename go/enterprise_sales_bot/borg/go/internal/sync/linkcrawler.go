package sync

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/ai"
	_ "modernc.org/sqlite"
)

var (
	scriptTagRegex   = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	styleTagRegex    = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	noscriptTagRegex = regexp.MustCompile(`(?is)<noscript[^>]*>.*?</noscript>`)
	navTagRegex      = regexp.MustCompile(`(?is)<nav[^>]*>.*?</nav>`)
	headerTagRegex   = regexp.MustCompile(`(?is)<header[^>]*>.*?</header>`)
	footerTagRegex   = regexp.MustCompile(`(?is)<footer[^>]*>.*?</footer>`)
	asideTagRegex    = regexp.MustCompile(`(?is)<aside[^>]*>.*?</aside>`)
	htmlTagRegex     = regexp.MustCompile(`(?is)<[^>]+>`)
	spaceRegex       = regexp.MustCompile(`\s+`)
)

const defaultOpenRouterFreeModel = "xiaomi/mimo-v2-flash:free"

type LinkClassifier func(ctx context.Context, title, description, content string) ([]string, error)

type LinkCrawlerOptions struct {
	Limit      int
	HTTPClient *http.Client
	Classifier LinkClassifier
}

type LinkCrawlerReport struct {
	Selected  int      `json:"selected"`
	Processed int      `json:"processed"`
	Succeeded int      `json:"succeeded"`
	Failed    int      `json:"failed"`
	Tagged    int      `json:"tagged"`
	Errors    []string `json:"errors"`
}

type pendingLink struct {
	UUID string
	URL  string
	Tags []string
}

func CrawlPendingLinks(ctx context.Context, dbPath string, opts LinkCrawlerOptions) (*LinkCrawlerReport, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 5
	}

	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	links, err := selectPendingLinks(ctx, db, limit)
	if err != nil {
		return nil, err
	}

	report := &LinkCrawlerReport{
		Selected: len(links),
		Errors:   []string{},
	}
	if len(links) == 0 {
		return report, nil
	}

	for _, link := range links {
		if err := markLinkResearchStatus(ctx, db, link.UUID, "running", nil, nil, nil, nil, nil, false); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("mark running %s: %v", link.UUID, err))
			continue
		}

		report.Processed++
		metadata, err := crawlLink(ctx, httpClient, link.URL)
		if err != nil {
			report.Failed++
			httpStatus := parseHTTPStatus(err)
			if markErr := markLinkResearchStatus(ctx, db, link.UUID, "failed", nil, nil, nil, httpStatus, nil, false); markErr != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("mark failed %s: %v", link.UUID, markErr))
			} else {
				report.Errors = append(report.Errors, fmt.Sprintf("crawl %s: %v", link.URL, err))
			}
			continue
		}

		tags := link.Tags
		tagged := false
		if len(tags) == 0 && strings.TrimSpace(metadata.Content) != "" && opts.Classifier != nil {
			classifiedTags, classifyErr := opts.Classifier(ctx, metadata.Title, metadata.Description, metadata.Content)
			if classifyErr != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("classify %s: %v", link.URL, classifyErr))
			} else if len(classifiedTags) > 0 {
				tags = classifiedTags
				tagged = true
				report.Tagged++
			}
		}

		if err := markLinkResearchStatus(
			ctx,
			db,
			link.UUID,
			"done",
			nullableString(metadata.Title),
			nullableString(metadata.Description),
			nullableString(metadata.FaviconURL),
			nullableInt(metadata.HTTPStatus),
			tags,
			tagged,
		); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("mark done %s: %v", link.UUID, err))
			continue
		}

		report.Succeeded++
	}

	return report, nil
}

type crawledLink struct {
	Title       string
	Description string
	FaviconURL  string
	Content     string
	HTTPStatus  int
}

func crawlLink(ctx context.Context, httpClient *http.Client, rawURL string) (*crawledLink, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, err
	}
	html := string(body)

	title := firstNonEmptyString(
		extractTagText(html, `(?is)<title[^>]*>(.*?)</title>`),
		extractMetaContent(html, "property", "og:title"),
	)
	description := firstNonEmptyString(
		extractMetaContent(html, "name", "description"),
		extractMetaContent(html, "property", "og:description"),
	)
	faviconURL := firstNonEmptyString(
		extractLinkHref(html, `(?is)<link[^>]+rel=["'][^"']*icon[^"']*["'][^>]+href=["']([^"']+)["'][^>]*>`),
		extractLinkHref(html, `(?is)<link[^>]+href=["']([^"']+)["'][^>]+rel=["'][^"']*icon[^"']*["'][^>]*>`),
	)
	content := extractVisibleText(html)
	if len(content) > 5000 {
		content = content[:5000]
	}

	return &crawledLink{
		Title:       title,
		Description: description,
		FaviconURL:  faviconURL,
		Content:     content,
		HTTPStatus:  resp.StatusCode,
	}, nil
}

func DefaultLinkTagClassifier(ctx context.Context, title, description, content string) ([]string, error) {
	response, err := ai.AutoRouteWithModel(ctx, defaultOpenRouterFreeModel, []ai.Message{
		{Role: "system", Content: "Extract concise webpage tags. Return valid JSON only with a single field named tags containing an array of 3 to 5 short strings."},
		{Role: "user", Content: fmt.Sprintf("Title: %s\nDescription: %s\nContent: %s", title, description, content)},
	})
	if err != nil {
		return nil, err
	}

	cleaned := strings.TrimSpace(response.Content)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	var payload struct {
		Tags []string `json:"tags"`
	}
	if err := json.Unmarshal([]byte(cleaned), &payload); err != nil {
		return nil, err
	}
	return normalizeTags(payload.Tags), nil
}

func normalizeTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		normalized := strings.TrimSpace(tag)
		if normalized == "" {
			continue
		}
		key := strings.ToLower(normalized)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, normalized)
		if len(result) >= 5 {
			break
		}
	}
	return result
}

func selectPendingLinks(ctx context.Context, db *sql.DB, limit int) ([]pendingLink, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT uuid, url, tags
		FROM links_backlog
		WHERE research_status = 'pending'
		ORDER BY created_at ASC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]pendingLink, 0, limit)
	for rows.Next() {
		var uuidValue string
		var urlValue string
		var tagsRaw sql.NullString
		if err := rows.Scan(&uuidValue, &urlValue, &tagsRaw); err != nil {
			return nil, err
		}
		result = append(result, pendingLink{
			UUID: uuidValue,
			URL:  urlValue,
			Tags: parseJSONStringArray(tagsRaw.String),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func markLinkResearchStatus(ctx context.Context, db *sql.DB, uuidValue, status string, pageTitle, pageDescription, faviconURL *string, httpStatus *int, tags []string, updateTags bool) error {
	timestamp := time.Now().UnixMilli()
	var tagsJSON any
	if updateTags {
		encoded, err := json.Marshal(tags)
		if err != nil {
			return err
		}
		tagsJSON = string(encoded)
	} else {
		tagsJSON = nil
	}

	_, err := db.ExecContext(ctx, `
		UPDATE links_backlog
		SET research_status = ?,
			http_status = COALESCE(?, http_status),
			page_title = COALESCE(?, page_title),
			page_description = COALESCE(?, page_description),
			favicon_url = COALESCE(?, favicon_url),
			tags = COALESCE(?, tags),
			researched_at = CASE WHEN ? = 'done' THEN ? ELSE researched_at END,
			updated_at = ?
		WHERE uuid = ?
	`, status, httpStatus, pageTitle, pageDescription, faviconURL, tagsJSON, status, timestamp, timestamp, uuidValue)
	return err
}

func parseJSONStringArray(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var parsed []string
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil
	}
	return normalizeTags(parsed)
}

func parseHTTPStatus(err error) *int {
	if err == nil {
		return nil
	}
	message := err.Error()
	if !strings.HasPrefix(message, "HTTP ") {
		return nil
	}
	statusCode, convErr := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(message, "HTTP ")))
	if convErr != nil {
		return nil
	}
	return &statusCode
}

func nullableString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func nullableInt(value int) *int {
	if value == 0 {
		return nil
	}
	return &value
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func extractTagText(input, pattern string) string {
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(input)
	if len(match) < 2 {
		return ""
	}
	return decodeHTMLWhitespace(match[1])
}

func extractMetaContent(input, attrName, attrValue string) string {
	patterns := []string{
		fmt.Sprintf(`(?is)<meta[^>]+%s=["']%s["'][^>]+content=["']([^"']+)["'][^>]*>`, regexp.QuoteMeta(attrName), regexp.QuoteMeta(attrValue)),
		fmt.Sprintf(`(?is)<meta[^>]+content=["']([^"']+)["'][^>]+%s=["']%s["'][^>]*>`, regexp.QuoteMeta(attrName), regexp.QuoteMeta(attrValue)),
	}
	for _, pattern := range patterns {
		if value := extractTagText(input, pattern); value != "" {
			return value
		}
	}
	return ""
}

func extractLinkHref(input, pattern string) string {
	return extractTagText(input, pattern)
}

func extractVisibleText(input string) string {
	// Strip standard non-visible blocks
	withoutScripts := scriptTagRegex.ReplaceAllString(input, " ")
	withoutStyles := styleTagRegex.ReplaceAllString(withoutScripts, " ")
	withoutNoscript := noscriptTagRegex.ReplaceAllString(withoutStyles, " ")

	// Phase 1 Fit Markdown: aggressively remove boilerplate nav/header/footer/sidebar elements
	withoutNav := navTagRegex.ReplaceAllString(withoutNoscript, " ")
	withoutHeader := headerTagRegex.ReplaceAllString(withoutNav, " ")
	withoutFooter := footerTagRegex.ReplaceAllString(withoutHeader, " ")
	withoutAside := asideTagRegex.ReplaceAllString(withoutFooter, " ")

	// Strip all remaining HTML tags
	withoutTags := htmlTagRegex.ReplaceAllString(withoutAside, " ")

	// Clean up massive whitespace blocks resulting from tag stripping
	collapsedSpaces := spaceRegex.ReplaceAllString(withoutTags, " ")

	return decodeHTMLWhitespace(collapsedSpaces)
}

func decodeHTMLWhitespace(input string) string {
	replacer := strings.NewReplacer(
		"&nbsp;", " ",
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", `"`,
		"&#39;", "'",
	)
	cleaned := replacer.Replace(input)
	return strings.Join(strings.Fields(cleaned), " ")
}
