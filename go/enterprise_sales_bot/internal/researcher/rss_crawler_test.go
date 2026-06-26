package researcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testRSSFeed = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
<title>Tech Blog</title>
<link>https://techblog.example.com</link>
<description>Latest tech news</description>
<item>
<title>Acme Corp is hiring senior ML engineers</title>
<link>https://techblog.example.com/acme-hiring</link>
<description>Acme Corp is looking for experienced machine learning engineers to join their AI team. Competitive salary and remote options available.</description>
<pubDate>Mon, 10 Jun 2026 12:00:00 +0000</pubDate>
</item>
<item>
<title>Building scalable infrastructure with Rust</title>
<link>https://techblog.example.com/rust-infra</link>
<description>A deep dive into how we migrated our microservices to Rust for better performance and safety.</description>
<pubDate>Mon, 03 Jun 2026 09:00:00 +0000</pubDate>
</item>
<item>
<title>Announcing our new open source LLM framework</title>
<link>https://techblog.example.com/llm-framework</link>
<description>We are excited to announce our new open source framework for building LLM-powered applications. Built on top of the latest advances in agent orchestration.</description>
<pubDate>Mon, 20 May 2026 14:00:00 +0000</pubDate>
</item>
</channel>
</rss>`

const testAtomFeed = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
<title>Engineering Blog</title>
<link href="https://engblog.example.com"/>
<updated>2026-06-14T12:00:00Z</updated>
<entry>
<title>GlobalTech Inc is hiring AI engineers</title>
<link href="https://engblog.example.com/globaltech-hiring"/>
<summary>GlobalTech Inc is expanding their AI division and hiring multiple engineers for their agentic workflow platform team.</summary>
<updated>2026-06-12T10:00:00Z</updated>
</entry>
<entry>
<title>How we reduced latency by 60%</title>
<link href="https://engblog.example.com/latency-reduction"/>
<summary>Our engineering team implemented a new caching strategy that dramatically reduced API latency across our platform.</summary>
<updated>2026-06-05T08:00:00Z</updated>
</entry>
</feed>`

func TestRSSFeedCrawler_Crawl_HiringSignalFound(t *testing.T) {
	// Start a test server serving an RSS feed
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(testRSSFeed))
	}))
	defer server.Close()

	crawler := &RSSFeedCrawler{
		Feeds: []string{server.URL},
		Keywords: []string{"hiring", "engineer"},
	}

	result, err := crawler.Crawl(context.Background(), "acme")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Fatal("expected non-empty result for hiring match")
	}
	if !strings.Contains(result, "Acme Corp is hiring") {
		t.Errorf("expected result to contain 'Acme Corp is hiring', got: %s", result)
	}
	if !strings.Contains(result, "techblog.example.com/acme-hiring") {
		t.Errorf("expected result to contain the article link, got: %s", result)
	}
}

func TestRSSFeedCrawler_Crawl_NoMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(testRSSFeed))
	}))
	defer server.Close()

	crawler := &RSSFeedCrawler{
		Feeds: []string{server.URL},
	}

	// Target that doesn't appear in the feed
	result, err := crawler.Crawl(context.Background(), "nonexistentcorp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty result, got: %s", result)
	}
}

func TestRSSFeedCrawler_Crawl_AtomFeed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/atom+xml")
		_, _ = w.Write([]byte(testAtomFeed))
	}))
	defer server.Close()

	crawler := &RSSFeedCrawler{
		Feeds: []string{server.URL},
	}

	result, err := crawler.Crawl(context.Background(), "globaltech")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Fatal("expected non-empty result for Atom hiring match")
	}
	if !strings.Contains(result, "GlobalTech Inc is hiring") {
		t.Errorf("expected result to contain 'GlobalTech Inc is hiring', got: %s", result)
	}
	if !strings.Contains(result, "engblog.example.com/globaltech-hiring") {
		t.Errorf("expected result to contain the article link, got: %s", result)
	}
}

func TestRSSFeedCrawler_Crawl_EmptyTarget(t *testing.T) {
	crawler := &RSSFeedCrawler{
		Feeds: []string{"https://example.com/feed.xml"},
	}

	result, err := crawler.Crawl(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty result for empty target, got: %s", result)
	}
}

func TestRSSFeedCrawler_Crawl_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	crawler := &RSSFeedCrawler{
		Feeds: []string{server.URL},
	}

	result, err := crawler.Crawl(context.Background(), "acme")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty result on HTTP error, got: %s", result)
	}
}

func TestRSSFeedCrawler_Crawl_InvalidXML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte("this is not valid xml"))
	}))
	defer server.Close()

	crawler := &RSSFeedCrawler{
		Feeds: []string{server.URL},
	}

	result, err := crawler.Crawl(context.Background(), "acme")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty result on invalid XML, got: %s", result)
	}
}

func TestContainsKeyword(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		keywords []string
		want     bool
	}{
		{"match single keyword", "we are hiring engineers", []string{"hiring", "budget"}, true},
		{"no match", "this is just a regular article", []string{"hiring", "job", "career"}, false},
		{"case insensitive", "We Are HIRING Now", []string{"hiring"}, true},
		{"empty keywords", "anything at all", []string{}, false},
		{"empty text", "", []string{"hiring"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsKeyword(tt.text, tt.keywords)
			if got != tt.want {
				t.Errorf("containsKeyword(%q, %v) = %v, want %v", tt.text, tt.keywords, got, tt.want)
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		name    string
		dateStr string
		wantErr bool
	}{
		{"RFC1123Z", "Mon, 10 Jun 2026 12:00:00 +0000", false},
		{"RFC1123", "Mon, 10 Jun 2026 12:00:00 UTC", false},
		{"RFC3339", "2026-06-10T12:00:00Z", false},
		{"ISO8601", "2026-06-10T12:00:00+00:00", false},
		{"invalid", "not a date", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseDate(tt.dateStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDate(%q) error = %v, wantErr = %v", tt.dateStr, err, tt.wantErr)
			}
		})
	}
}
