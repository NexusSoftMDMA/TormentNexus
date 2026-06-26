package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// HNWhoIsHiringSource scrapes Hacker News "Ask HN: Who is hiring?" threads
// for companies posting AI/LLM/engineering roles. This is the highest-intent
// free signal available — companies literally advertising their pain points.
type HNWhoIsHiringSource struct {
	Client *http.Client
}

// hnItem represents a Hacker News API item.
type hnItem struct {
	ID		int	`json:"id"`
	By		string	`json:"by"`
	Text		string	`json:"text"`
	Kids		[]int	`json:"kids"`
	Time		int	`json:"time"`
	Type		string	`json:"type"`
	Title		string	`json:"title"`
	Descendants	int	`json:"descendants"`
}

// hnUser represents a Hacker News user profile.

var (
	// Matches company name from the first line of a HN hiring comment.
	// Format is typically: "Company Name | Role | Location | ..."
	hnCompanyRegex	= regexp.MustCompile(`^([^|]+?)(?:\s*\||\s*$)`)

	// Matches URLs in the comment text (company websites).
	hnURLRegex	= regexp.MustCompile(`https?://[^\s<>"]+`)

	// Matches domain from URL.
	hnDomainRegex	= regexp.MustCompile(`(?:https?://)?(?:www\.)?([^/\s]+)`)

	// Keywords that signal AI/LLM relevance.
	aiKeywords	= []string{
		"llm", "large language model", "ai engineer", "machine learning",
		"ml engineer", "deep learning", "nlp", "natural language",
		"generative ai", "gen ai", "agentic", "agent", "orchestration",
		"rag", "retrieval augmented", "vector", "embedding",
		"transformer", "gpt", "claude", "gemini", "openai",
		"anthropic", "mlops", "ml platform", "ai platform",
		"ai infrastructure", "inference", "fine-tuning", "fine tuning",
		"prompt engineer", "langchain", "llamaindex",
	}
)

const (
	hnAPIBase	= "https://hacker-news.firebaseio.com/v0"
	hnSearchBase	= "https://hn.algolia.com/api/v1"
)

// Discover implements LeadSource. It searches HN "Who is Hiring" threads
// for companies with AI/LLM-related job postings.
func (h *HNWhoIsHiringSource) Discover(ctx context.Context, keywords []string) ([]db.Company, error) {
	slog.Info("HNWhoIsHiring: Searching for AI/LLM hiring signals...")

	// 1. Find the latest "Who is Hiring" thread
	threadID, err := h.findLatestWhoIsHiringThread(ctx)
	if err != nil {
		return nil, fmt.Errorf("hn: failed to find hiring thread: %w", err)
	}

	if threadID == 0 {
		slog.Info("HNWhoIsHiring: No active 'Who is Hiring' thread found")
		return nil, nil
	}

	slog.Info(fmt.Sprintf("HNWhoIsHiring: Found thread https://news.ycombinator.com/item?id=%d", threadID))

	// 2. Fetch the thread's top-level comments (these are job postings)
	comments, err := h.fetchThreadComments(ctx, threadID)
	if err != nil {
		return nil, fmt.Errorf("hn: failed to fetch comments: %w", err)
	}

	slog.Info(fmt.Sprintf("HNWhoIsHiring: Found %d top-level comments, filtering for AI/LLM...", len(comments)))

	// 3. Parse each comment for company info and AI relevance
	var companies []db.Company
	seen := make(map[string]bool)	// deduplicate by domain

	for _, comment := range comments {
		if comment.Text == "" {
			continue
		}

		company, ok := h.parseComment(comment)
		if !ok {
			continue
		}

		// Deduplicate by domain
		if seen[company.Domain] {
			continue
		}

		// Check AI relevance — either from keywords or user-supplied
		if h.isAIRelevant(comment.Text, keywords) {
			seen[company.Domain] = true
			companies = append(companies, company)
			slog.Info(fmt.Sprintf("HNWhoIsHiring: Found AI-relevant company: %s (%s)", company.Name, company.Domain))
		}
	}

	slog.Info(fmt.Sprintf("HNWhoIsHiring: Discovered %d AI-relevant companies from thread", len(companies)))
	return companies, nil
}

// findLatestWhoIsHiringThread uses the Algolia HN search API to find the
// most recent "Who is Hiring" thread by whoishiring.
func (h *HNWhoIsHiringSource) findLatestWhoIsHiringThread(ctx context.Context) (int, error) {
	// Use search_by_date to get results sorted by newest first
	url := fmt.Sprintf("%s/search_by_date?query=%%22who+is+hiring%%22&tags=story,author_whoishiring&hitsPerPage=5&numericFilters=created_at_i>%d",
		hnSearchBase, time.Now().AddDate(0, -3, 0).Unix())	// within last 3 months

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "TormentNexus-SalesBot/1.0 (contact: pelloni.robert@gmail.com)")

	resp, err := h.client().Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		Hits []struct {
			ObjectID	string	`json:"objectID"`
			Title		string	`json:"title"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	for _, hit := range result.Hits {
		title := strings.ToLower(hit.Title)
		if strings.Contains(title, "who is hiring") && !strings.Contains(title, "freelancer") {
			var id int
			if _, err := fmt.Sscanf(hit.ObjectID, "%d", &id); err == nil && id > 0 {
				return id, nil
			}
		}
	}

	return 0, nil
}

// fetchThreadComments fetches all top-level comments from a HN thread.
// Uses the HN API item endpoint and fetches kids in parallel.
func (h *HNWhoIsHiringSource) fetchThreadComments(ctx context.Context, threadID int) ([]hnItem, error) {
	// First, get the thread item to find kid IDs
	thread, err := h.fetchItem(ctx, threadID)
	if err != nil {
		return nil, err
	}

	if len(thread.Kids) == 0 {
		return nil, nil
	}

	// Fetch top-level comments (limit to first 200 for performance)
	limit := len(thread.Kids)
	if limit > 200 {
		limit = 200
	}

	comments := make([]hnItem, 0, limit)
	type result struct {
		item	hnItem
		err	error
	}

	ch := make(chan result, limit)
	sem := make(chan struct{}, 10)	// concurrency limit

	for _, kidID := range thread.Kids[:limit] {
		sem <- struct{}{}
		go func(id int) {
			defer func() { <-sem }()
			item, err := h.fetchItem(ctx, id)
			ch <- result{item: item, err: err}
		}(kidID)
	}

	for i := 0; i < limit; i++ {
		r := <-ch
		if r.err == nil && r.item.Type == "comment" && r.item.Text != "" {
			comments = append(comments, r.item)
		}
	}

	return comments, nil
}

// fetchItem fetches a single HN item by ID.
func (h *HNWhoIsHiringSource) fetchItem(ctx context.Context, id int) (hnItem, error) {
	url := fmt.Sprintf("%s/item/%d.json", hnAPIBase, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return hnItem{}, err
	}
	req.Header.Set("User-Agent", "TormentNexus-SalesBot/1.0 (contact: pelloni.robert@gmail.com)")

	resp, err := h.client().Do(req)
	if err != nil {
		return hnItem{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	var item hnItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return hnItem{}, err
	}

	return item, nil
}

// parseComment extracts company name and domain from a HN hiring comment.
// The first line of most "Who is Hiring" comments follows: "Company | Role | Location | ..."
func (h *HNWhoIsHiringSource) parseComment(comment hnItem) (db.Company, bool) {
	// Strip HTML tags from HN comment text
	text := stripHTML(comment.Text)
	lines := strings.SplitN(text, "\n", 2)
	if len(lines) == 0 {
		return db.Company{}, false
	}

	firstLine := strings.TrimSpace(lines[0])
	if len(firstLine) < 3 {
		return db.Company{}, false
	}

	// Extract company name from first line
	matches := hnCompanyRegex.FindStringSubmatch(firstLine)
	if len(matches) < 2 {
		return db.Company{}, false
	}

	companyName := strings.TrimSpace(matches[1])
	if len(companyName) < 2 || len(companyName) > 100 {
		return db.Company{}, false
	}

	// Try to extract domain from URLs in the comment
	domain := ""
	urls := hnURLRegex.FindAllString(text, 5)
	for _, u := range urls {
		dm := hnDomainRegex.FindStringSubmatch(u)
		if len(dm) > 1 {
			d := strings.ToLower(dm[1])
			// Skip common non-company domains
			if !isExcludedDomain(d) {
				domain = d
				break
			}
		}
	}

	// If no URL found, generate a placeholder domain from company name
	if domain == "" {
		domain = strings.ToLower(companyName)
		domain = regexp.MustCompile(`[^a-z0-9]`).ReplaceAllString(domain, "")
		if len(domain) > 30 {
			domain = domain[:30]
		}
		domain += ".com"
	}

	// Extract tech stack from comment text
	techStack := extractTechStack(text)

	// Build hiring signals from the comment
	hiringSignals := []string{firstLine}
	if len(lines) > 1 && len(lines[1]) > 10 {
		// Add second line as additional context (usually role description)
		summary := lines[1]
		if len(summary) > 200 {
			summary = summary[:200] + "..."
		}
		hiringSignals = append(hiringSignals, summary)
	}

	return db.Company{
		Name:		companyName,
		Domain:		domain,
		TechStack:	techStack,
		HiringSignals:	hiringSignals,
		MarketCapTier:	classifyMarketCap(text),
	}, true
}

// isAIRelevant checks if a comment text contains AI/LLM-related keywords.
func (h *HNWhoIsHiringSource) isAIRelevant(text string, extraKeywords []string) bool {
	lower := strings.ToLower(text)

	// Check built-in AI keywords
	for _, kw := range aiKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}

	// Check user-supplied keywords
	for _, kw := range extraKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}

	return false
}

// extractTechStack parses common tech keywords from job posting text.
func extractTechStack(text string) []string {
	lower := strings.ToLower(text)
	var stack []string

	techMap := map[string]string{
		"go":		"Go",
		"golang":	"Go",
		"python":	"Python",
		"rust":		"Rust",
		"typescript":	"TypeScript",
		"kubernetes":	"Kubernetes",
		"k8s":		"Kubernetes",
		"docker":	"Docker",
		"aws":		"AWS",
		"gcp":		"GCP",
		"azure":	"Azure",
		"postgres":	"PostgreSQL",
		"postgresql":	"PostgreSQL",
		"redis":	"Redis",
		"kafka":	"Kafka",
		"pytorch":	"PyTorch",
		"tensorflow":	"TensorFlow",
		"langchain":	"LangChain",
		"llamaindex":	"LlamaIndex",
		"openai":	"OpenAI",
		"anthropic":	"Anthropic",
		"triton":	"Triton",
		"vllm":		"vLLM",
		"ray":		"Ray",
	}

	seen := make(map[string]bool)
	for keyword, display := range techMap {
		if strings.Contains(lower, keyword) && !seen[display] {
			seen[display] = true
			stack = append(stack, display)
		}
	}

	return stack
}

// classifyMarketCap estimates company tier from context clues in the posting.
func classifyMarketCap(text string) string {
	lower := strings.ToLower(text)

	if strings.Contains(lower, "startup") || strings.Contains(lower, "seed") || strings.Contains(lower, "series a") {
		return "Startup"
	}
	if strings.Contains(lower, "enterprise") || strings.Contains(lower, "fortune") || strings.Contains(lower, "public company") {
		return "Enterprise"
	}
	if strings.Contains(lower, "yc ") || strings.Contains(lower, "y combinator") {
		return "Startup"
	}

	return "Mid-Market"
}

// isExcludedDomain filters out common non-company domains.
func isExcludedDomain(domain string) bool {
	excluded := []string{
		"github.com", "linkedin.com", "twitter.com", "x.com",
		"youtube.com", "medium.com", "substack.com", "notion.site",
		"notion.so", "google.com", "apple.com", "microsoft.com",
		"amazon.com", "facebook.com", "meta.com", "greenhouse.io",
		"lever.co", "workday.com", "ashbyhq.com", "jobs.lever.co",
		"boards.greenhouse.io", "apply.workable.com",
		"hn.algolia.com", "news.ycombinator.com", "ycombinator.com",
		"docs.google.com", "forms.gle",
	}

	for _, ex := range excluded {
		if strings.HasSuffix(domain, ex) {
			return true
		}
	}

	return false
}

// stripHTML removes HTML tags from HN comment text.
func stripHTML(s string) string {
	// Replace <p> with newlines
	s = strings.ReplaceAll(s, "<p>", "\n")
	// Replace <br> with newlines
	s = strings.ReplaceAll(s, "<br>", "\n")
	s = strings.ReplaceAll(s, "<br/>", "\n")
	s = strings.ReplaceAll(s, "<br />", "\n")
	// Remove remaining HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	s = re.ReplaceAllString(s, "")
	// Decode common HTML entities
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#x27;", "'")
	s = strings.ReplaceAll(s, "&apos;", "'")
	// Clean up whitespace
	s = strings.TrimSpace(s)
	return s
}

func (h *HNWhoIsHiringSource) client() *http.Client {
	if h.Client != nil {
		return h.Client
	}
	return &http.Client{Timeout: 30 * time.Second}
}
