//go:build ignore
// +build ignore

package tools

/**
 * @file octagon.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Octagon financial intelligence.
 * Replaces `octagon` (npx octagon-mcp) and `octagon-deep-research`
 * (npx octagon-deep-research-mcp) STDIO entries in mcp.json.
 *
 * Uses the Octagon REST API (https://api.octagon.dev).
 * Improvements over original:
 *  - No npx/Node dependency.
 *  - Unified tool: company search, financial data, deep research, news.
 *  - Context-aware with timeout; requires OCTAGON_API_KEY.
 */

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const octagonBaseURL = "https://api.octagon.dev/v1"

func octagonAPIKey() string {
	return os.Getenv("OCTAGON_API_KEY")
}

func octagonDo(ctx context.Context, method, path string, payload interface{}) (interface{}, error) {
	apiKey := octagonAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("OCTAGON_API_KEY environment variable is not set")
	}

	var bodyReader io.Reader
	if payload != nil {
		data, _ := json.Marshal(payload)
		bodyReader = bytes.NewBuffer(data)
	}

	req, e := http.NewRequestWithContext(ctx, method, octagonBaseURL+path, bodyReader)
	if e != nil {
		return nil, e
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, e := client.Do(req)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Octagon API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return string(body), nil
	}
	return result, nil
}

// HandleOctagonResearch performs deep company or topic research via Octagon.
// Tool: octagon_research
func HandleOctagonResearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "topic", "company")
	if query == "" {
		return err("query parameter is required")
	}

	payload := map[string]interface{}{"query": query}

	if depth, _ := getString(args, "depth"); depth != "" {
		payload["depth"] = depth
	}

	result, e := octagonDo(ctx, "POST", "/research", payload)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleOctagonCompanySearch searches for companies in Octagon's database.
// Tool: octagon_company_search
func HandleOctagonCompanySearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q", "company")
	if query == "" {
		return err("query parameter is required")
	}

	limit := getInt(args, "limit", "count")
	if limit <= 0 {
		limit = 10
	}

	result, e := octagonDo(ctx, "POST", "/companies/search", map[string]interface{}{
		"query": query,
		"limit": limit,
	})
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleOctagonFinancials retrieves financial data for a company.
// Tool: octagon_financials
func HandleOctagonFinancials(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	company, _ := getString(args, "company", "ticker", "symbol", "name")
	if company == "" {
		return err("company parameter is required")
	}

	dataType, _ := getString(args, "type", "data_type")
	if dataType == "" {
		dataType = "summary"
	}

	result, e := octagonDo(ctx, "POST", "/financials", map[string]interface{}{
		"company": company,
		"type":    dataType,
	})
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleOctagonNews retrieves news and intelligence about a company/topic.
// Tool: octagon_news
func HandleOctagonNews(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "company", "topic")
	if query == "" {
		return err("query parameter is required")
	}

	limit := getInt(args, "limit", "count")
	if limit <= 0 {
		limit = 10
	}

	result, e := octagonDo(ctx, "POST", "/news", map[string]interface{}{
		"query": query,
		"limit": limit,
	})
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}
