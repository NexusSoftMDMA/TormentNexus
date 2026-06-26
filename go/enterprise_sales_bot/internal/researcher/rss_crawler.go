package researcher

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// RSSItem represents a single entry in an RSS 2.0 feed.
type RSSItem struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
}

type RSSChannel struct {
	Items []RSSItem `xml:"item"`
}

type RSS struct {
	Channel RSSChannel `xml:"channel"`
}

// AtomEntry represents a single entry in an Atom feed.
type AtomEntry struct {
	Title   string `xml:"title"`
	Summary string `xml:"summary"`
	Link    struct {
		Href string `xml:"href,attr"`
	} `xml:"link"`
	Updated string `xml:"updated"`
}

type Atom struct {
	Entries []AtomEntry `xml:"entry"`
}

// RSSFeedCrawler scans configured RSS/Atom feeds for hiring signals and company mentions.
// It implements the researcher.Crawler interface.
type RSSFeedCrawler struct {
	// HTTP client used for fetching feeds. If nil, http.DefaultClient is used.
	Client *http.Client
	// Feeds is a list of RSS/Atom feed URLs to scan.
	Feeds []string
	// Keywords used to detect hiring or interest signals.
	Keywords []string
	// MaxAge limits how old an article can be to be considered (e.g., 30 days).
	MaxAge time.Duration
}

// defaultKeywords returns a set of terms that commonly indicate hiring or technical interest.
func defaultKeywords() []string {
	return []string{"hiring", "engineer", "ml", "ai", "machine learning", "data science", "open position", "job", "career", "vacancy", "join us", "team", "role"}
}

// Crawl implements the Crawler interface. It scans all configured feeds, looking for
// any article that mentions the target (company domain or name) and any of the hiring keywords.
// It returns a simple concatenated string of findings, or an empty string if none are found.
func (r *RSSFeedCrawler) Crawl(ctx context.Context, target string) (string, error) {
	if target == "" {
		return "", nil
	}
	if r.Client == nil {
		r.Client = http.DefaultClient
	}
	// Use supplied keywords if set, otherwise defaults.
	keywords := r.Keywords
	if len(keywords) == 0 {
		keywords = defaultKeywords()
	}
	// MaxAge defaults to 30 days if not set.
	maxAge := r.MaxAge
	if maxAge == 0 {
		maxAge = 30 * 24 * time.Hour
	}

	// Prepare for deduplication of matching items.
	var matches []string
	seen := make(map[string]struct{})
	targetLower := strings.ToLower(target)

	for _, feedURL := range r.Feeds {
		req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
		if err != nil {
			slog.Info(fmt.Sprintf("RSSFeedCrawler: failed to create request for %s: %v", feedURL, err))
			continue
		}
		resp, err := r.Client.Do(req)
		if err != nil {
			slog.Info(fmt.Sprintf("RSSFeedCrawler: request error for %s: %v", feedURL, err))
			continue
		}
		if resp.StatusCode != http.StatusOK {
			slog.Info(fmt.Sprintf("RSSFeedCrawler: non-200 status %d for %s", resp.StatusCode, feedURL))
			_ = resp.Body.Close()
			continue
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			slog.Info(fmt.Sprintf("RSSFeedCrawler: read error for %s: %v", feedURL, err))
			continue
		}

		// Attempt RSS parsing first.
		var rss RSS
		if err := xml.Unmarshal(body, &rss); err == nil && len(rss.Channel.Items) > 0 {
			for _, item := range rss.Channel.Items {
				// Combine title & description for matching.
				content := strings.ToLower(item.Title + " " + item.Description)
				// Check article age.
				if item.PubDate != "" {
					if t, err := parseDate(item.PubDate); err == nil && time.Since(t) > maxAge {
						continue
					}
				}
				if strings.Contains(content, targetLower) && containsKeyword(content, keywords) {
					// Deduplicate by link.
					if _, ok := seen[item.Link]; ok {
						continue
					}
					seen[item.Link] = struct{}{}
					matches = append(matches, fmt.Sprintf("RSS: %s – %s", item.Title, item.Link))
				}
			}
			continue // RSS parsed successfully; skip Atom attempt.
		}

		// Attempt Atom parsing.
		var atom Atom
		if err := xml.Unmarshal(body, &atom); err == nil && len(atom.Entries) > 0 {
			for _, entry := range atom.Entries {
				content := strings.ToLower(entry.Title + " " + entry.Summary)
				// Check article age using Updated field.
				if entry.Updated != "" {
					if t, err := parseDate(entry.Updated); err == nil && time.Since(t) > maxAge {
						continue
					}
				}
				if strings.Contains(content, targetLower) && containsKeyword(content, keywords) {
					link := entry.Link.Href
					if _, ok := seen[link]; ok {
						continue
					}
					seen[link] = struct{}{}
					matches = append(matches, fmt.Sprintf("Atom: %s – %s", entry.Title, link))
				}
			}
		}
	}

	if len(matches) == 0 {
		return "", nil
	}
	// Join findings with newlines.
	return strings.Join(matches, "\n"), nil
}

// containsKeyword reports true if any of the supplied keywords appear in the text.
func containsKeyword(text string, kw []string) bool {
	textLower := strings.ToLower(text)
	for _, k := range kw {
		if strings.Contains(textLower, strings.ToLower(k)) {
			return true
		}
	}
	return false
}

// parseDate attempts to parse common RSS/Atom date formats.
func parseDate(dateStr string) (time.Time, error) {
	// Try multiple known formats.
	layouts := []string{
		time.RFC1123Z, // "Mon, 02 Jan 2006 15:04:05 -0700"
		time.RFC1123,  // "Mon, 02 Jan 2006 15:04:05 MST"
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00", // fallback same as RFC3339
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, nil
		}
	}
	// If parsing fails, return zero time with error.
	return time.Time{}, fmt.Errorf("unable to parse date %s", dateStr)
}
