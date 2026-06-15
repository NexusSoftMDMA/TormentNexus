//go:build ignore
// +build ignore

package tools

/**
 * @file google_news.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Google News MCP server.
 * Replaces: server-google-news (npm)
 *
 * Provides Google News search and headlines via RSS/XML feeds.
 * No API key required.
 *
 * Tools:
 *  - google_news_headlines — get top headlines
 *  - google_news_search — search Google News
 */

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const gNewsRSS = "https://news.google.com/rss"

// RSS structures for parsing
type rssFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Description string    `xml:"description"`
		Items       []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	Source      string `xml:"source"`
	Description string `xml:"description"`
}

func fetchGoogleRSS(ctx context.Context, path string) ([]rssItem, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Get(gNewsRSS + path)
	if e != nil {
		return nil, fmt.Errorf("fetch failed: %v", e)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	var feed rssFeed
	if e := xml.Unmarshal(data, &feed); e != nil {
		return nil, fmt.Errorf("parse failed: %v", e)
	}
	return feed.Channel.Items, nil
}

// HandleGoogleNewsHeadlines returns top news headlines.
func HandleGoogleNewsHeadlines(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	category, _ := getString(args, "category", "topic")
	locale, _ := getString(args, "locale", "lang")

	path := ""
	if category != "" {
		path = "/headlines/section/topic/" + url.PathEscape(category)
	}
	if locale != "" {
		if path == "" {
			path = "/headlines"
		}
		path += "?hl=" + url.QueryEscape(locale)
	}
	if path == "" {
		path = "/headlines"
	}

	items, e := fetchGoogleRSS(ctx, path)
	if e != nil {
		return err(fmt.Sprintf("headlines failed: %v", e))
	}

	result := map[string]interface{}{
		"count":   len(items),
		"headlines": items,
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}

// HandleGoogleNewsSearch searches Google News.
func HandleGoogleNewsSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q", "search")
	if query == "" {
		return err("query is required")
	}
	locale, _ := getString(args, "locale", "lang")

	path := "/search?q=" + url.QueryEscape(query)
	if locale != "" {
		path += "&hl=" + url.QueryEscape(locale)
	}

	items, e := fetchGoogleRSS(ctx, path)
	if e != nil {
		return err(fmt.Sprintf("search failed: %v", e))
	}

	result := map[string]interface{}{
		"count":   len(items),
		"results": items,
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}
