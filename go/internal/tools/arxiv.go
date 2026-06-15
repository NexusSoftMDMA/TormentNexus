//go:build ignore
// +build ignore

package tools

/**
 * @file arxiv.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of arXiv academic paper search & retrieval.
 * Replaces the `arxiv-mcp-server` STDIO entry in mcp.json.
 *
 * Uses the public arXiv API (no key required).
 * Improvements over original:
 *  - No uvx/Python dependency.
 *  - Supports search, paper lookup by ID, and abstract extraction.
 *  - Context-aware with timeout.
 */

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const arxivBaseURL = "https://export.arxiv.org/api/query"

// arxivEntry represents a parsed arXiv feed entry.
type arxivEntry struct {
	ID        string   `xml:"id"`
	Updated   string   `xml:"updated"`
	Published string   `xml:"published"`
	Title     string   `xml:"title"`
	Summary   string   `xml:"summary"`
	Authors   []struct {
		Name string `xml:"name"`
	} `xml:"author"`
	Links []struct {
		Href string `xml:"href,attr"`
		Type string `xml:"type,attr"`
		Rel  string `xml:"rel,attr"`
	} `xml:"link"`
}

type arxivFeed struct {
	Entries []arxivEntry `xml:"entry"`
	Total   string       `xml:"totalResults"`
}

func arxivFetch(ctx context.Context, params map[string]string) ([]arxivEntry, error) {
	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, e := http.NewRequestWithContext(ctx, "GET", arxivBaseURL+"?"+q.Encode(), nil)
	if e != nil {
		return nil, e
	}
	req.Header.Set("User-Agent", "TormentNexus/1.0")

	resp, e := client.Do(req)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var feed arxivFeed
	if e := xml.Unmarshal(body, &feed); e != nil {
		return nil, fmt.Errorf("failed to parse arXiv response: %v", e)
	}
	return feed.Entries, nil
}

func formatEntry(e arxivEntry) string {
	authors := ""
	for i, a := range e.Authors {
		if i > 0 {
			authors += ", "
		}
		authors += a.Name
		if i >= 4 {
			authors += " et al."
			break
		}
	}

	pdfURL := ""
	for _, l := range e.Links {
		if l.Type == "application/pdf" || l.Rel == "related" {
			pdfURL = l.Href
			break
		}
	}

	title := e.Title
	summary := e.Summary
	if len(summary) > 1000 {
		summary = summary[:1000] + "..."
	}

	return fmt.Sprintf("**%s**\nID: %s\nAuthors: %s\nPublished: %s\nPDF: %s\n\n%s",
		title, e.ID, authors, e.Published, pdfURL, summary)
}

// HandleArxivSearch searches arXiv for academic papers.
// Tool: arxiv_search
func HandleArxivSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q", "search_query")
	if query == "" {
		return err("query parameter is required")
	}

	maxResults := getInt(args, "max_results", "maxResults", "count")
	if maxResults <= 0 {
		maxResults = 10
	}
	if maxResults > 50 {
		maxResults = 50
	}

	sortBy := "relevance"
	if sb, _ := getString(args, "sort_by", "sortBy"); sb != "" {
		sortBy = sb
	}

	sortOrder := "descending"
	if so, _ := getString(args, "sort_order", "sortOrder"); so != "" {
		sortOrder = so
	}

	params := map[string]string{
		"search_query": "all:" + query,
		"max_results":  fmt.Sprintf("%d", maxResults),
		"sortBy":       sortBy,
		"sortOrder":    sortOrder,
	}

	// Support category filter (e.g., "cs.AI", "physics.quant-ph")
	if cat, _ := getString(args, "category", "cat"); cat != "" {
		params["search_query"] = fmt.Sprintf("cat:%s AND all:%s", cat, query)
	}

	entries, e := arxivFetch(ctx, params)
	if e != nil {
		return err(e.Error())
	}

	if len(entries) == 0 {
		return ok("No arXiv papers found for query: " + query)
	}

	result := fmt.Sprintf("# arXiv Search Results for: %s\n\nFound %d papers:\n\n", query, len(entries))
	for i, entry := range entries {
		result += fmt.Sprintf("## %d. %s\n\n", i+1, formatEntry(entry))
	}
	return ok(result)
}

// HandleArxivGetPaper retrieves a specific arXiv paper by ID.
// Tool: arxiv_get_paper
func HandleArxivGetPaper(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	paperID, _ := getString(args, "paper_id", "id", "arxiv_id")
	if paperID == "" {
		return err("paper_id parameter is required (e.g., '2301.07041')")
	}

	params := map[string]string{
		"id_list": paperID,
	}

	entries, e := arxivFetch(ctx, params)
	if e != nil {
		return err(e.Error())
	}

	if len(entries) == 0 {
		return ok("Paper not found: " + paperID)
	}

	return ok(formatEntry(entries[0]))
}

// HandleArxivListRecent lists the most recent papers in a category.
// Tool: arxiv_list_recent
func HandleArxivListRecent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cat, _ := getString(args, "category", "cat")
	if cat == "" {
		cat = "cs.AI"
	}

	maxResults := getInt(args, "max_results", "maxResults", "count")
	if maxResults <= 0 {
		maxResults = 10
	}

	params := map[string]string{
		"search_query": "cat:" + cat,
		"max_results":  fmt.Sprintf("%d", maxResults),
		"sortBy":       "submittedDate",
		"sortOrder":    "descending",
	}

	entries, e := arxivFetch(ctx, params)
	if e != nil {
		return err(e.Error())
	}

	if len(entries) == 0 {
		return ok("No recent papers found in category: " + cat)
	}

	result := fmt.Sprintf("# Recent arXiv Papers in %s\n\n", cat)
	for i, entry := range entries {
		result += fmt.Sprintf("## %d. %s\n\n", i+1, formatEntry(entry))
	}
	return ok(result)
}
