//go:build ignore
// +build ignore

package tools

/**
 * @file semantic_scholar.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Semantic Scholar academic search.
 * Replaces `paper_search_server` (uvx paper-search-mcp) in mcp.json.
 *
 * Uses the public Semantic Scholar Academic Graph API.
 * Improvements over original:
 *  - No uvx/Python dependency.
 *  - Supports paper search, paper details, author lookup, citation graph.
 *  - Context-aware with timeout; optional API key for higher rate limits.
 */

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const s2BaseURL = "https://api.semanticscholar.org/graph/v1"

func s2Get(ctx context.Context, path string, params map[string]string) (map[string]interface{}, error) {
	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}

	fullURL := s2BaseURL + path
	if len(q) > 0 {
		fullURL += "?" + q.Encode()
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, e := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if e != nil {
		return nil, e
	}

	req.Header.Set("User-Agent", "TormentNexus/1.0")

	if apiKey := os.Getenv("SEMANTIC_SCHOLAR_API_KEY"); apiKey != "" {
		req.Header.Set("x-api-key", apiKey)
	}

	resp, e := client.Do(req)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Semantic Scholar API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return nil, fmt.Errorf("failed to parse Semantic Scholar response: %v", e)
	}
	return result, nil
}

func formatS2Paper(paper map[string]interface{}) string {
	title, _ := paper["title"].(string)
	year, _ := paper["year"].(float64)
	abstract, _ := paper["abstract"].(string)
	paperId, _ := paper["paperId"].(string)
	citationCount, _ := paper["citationCount"].(float64)
	isOpenAccess, _ := paper["isOpenAccess"].(bool)

	authors := ""
	if authorList, ok := paper["authors"].([]interface{}); ok {
		names := []string{}
		for _, a := range authorList {
			if aMap, ok := a.(map[string]interface{}); ok {
				if name, ok := aMap["name"].(string); ok {
					names = append(names, name)
				}
			}
		}
		if len(names) > 5 {
			authors = strings.Join(names[:5], ", ") + " et al."
		} else {
			authors = strings.Join(names, ", ")
		}
	}

	openAccessStr := ""
	if isOpenAccess {
		openAccessStr = " [Open Access]"
	}

	if len(abstract) > 800 {
		abstract = abstract[:800] + "..."
	}

	return fmt.Sprintf("**%s** (%d)%s\nID: %s\nAuthors: %s\nCitations: %d\n\n%s",
		title, int(year), openAccessStr, paperId, authors, int(citationCount), abstract)
}

// HandleSemanticScholarSearch searches for academic papers on Semantic Scholar.
// Tool: paper_search
func HandleSemanticScholarSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q")
	if query == "" {
		return err("query parameter is required")
	}

	limit := getInt(args, "limit", "count", "max_results")
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	fields := "paperId,title,abstract,authors,year,citationCount,isOpenAccess,openAccessPdf"

	params := map[string]string{
		"query":  query,
		"limit":  fmt.Sprintf("%d", limit),
		"fields": fields,
	}

	if year, _ := getString(args, "year"); year != "" {
		params["year"] = year
	}

	if fieldsOfStudy, _ := getString(args, "fieldsOfStudy"); fieldsOfStudy != "" {
		params["fieldsOfStudy"] = fieldsOfStudy
	}

	result, e := s2Get(ctx, "/paper/search", params)
	if e != nil {
		return err(e.Error())
	}

	papers, found := result["data"].([]interface{})
	if !found || len(papers) == 0 {
		return ok("No papers found for query: " + query)
	}

	out := fmt.Sprintf("# Semantic Scholar Search: %s\n\nFound %d papers:\n\n", query, len(papers))
	for i, p := range papers {
		if pMap, ok := p.(map[string]interface{}); ok {
			out += fmt.Sprintf("## %d. %s\n\n", i+1, formatS2Paper(pMap))
		}
	}
	return ok(out)
}

// HandleSemanticScholarGetPaper retrieves details for a specific paper by ID.
// Tool: paper_details
func HandleSemanticScholarGetPaper(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	paperID, _ := getString(args, "paper_id", "id")
	if paperID == "" {
		return err("paper_id parameter is required")
	}

	fields := "paperId,title,abstract,authors,year,citationCount,referenceCount,isOpenAccess,openAccessPdf,tldr"
	result, e := s2Get(ctx, "/paper/"+paperID, map[string]string{"fields": fields})
	if e != nil {
		return err(e.Error())
	}

	return ok(formatS2Paper(result))
}

// HandleSemanticScholarGetCitations retrieves citations for a paper.
// Tool: paper_citations
func HandleSemanticScholarGetCitations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	paperID, _ := getString(args, "paper_id", "id")
	if paperID == "" {
		return err("paper_id parameter is required")
	}

	limit := getInt(args, "limit", "count")
	if limit <= 0 {
		limit = 20
	}

	fields := "title,authors,year,citationCount"
	result, e := s2Get(ctx, fmt.Sprintf("/paper/%s/citations", paperID),
		map[string]string{"fields": fields, "limit": fmt.Sprintf("%d", limit)})
	if e != nil {
		return err(e.Error())
	}

	data, found := result["data"].([]interface{})
	if !found || len(data) == 0 {
		return ok("No citations found for paper: " + paperID)
	}

	out := fmt.Sprintf("# Citations for %s (%d total)\n\n", paperID, len(data))
	for i, c := range data {
		if cMap, ok := c.(map[string]interface{}); ok {
			if citing, ok := cMap["citingPaper"].(map[string]interface{}); ok {
				out += fmt.Sprintf("%d. %s\n\n", i+1, formatS2Paper(citing))
			}
		}
	}
	return ok(out)
}




