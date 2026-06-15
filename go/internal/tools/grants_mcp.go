//go:build ignore
// +build ignore

package tools

/**
 * @file grants_mcp.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Grants MCP — government grants discovery.
 * Replaces: github.com/tar-ive/grants-mcp
 *
 * Provides government grants discovery and analysis via the SimplersGrants.gov API.
 * No API key required (public data).
 *
 * Tools:
 *  - grants_search — search grant opportunities by keyword
 *  - grants_by_agency — search grants by agency
 *  - grants_by_category — search grants by funding category
 *  - grants_trends — analyze funding trends
 */

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const grantsAPIBase = "https://api.grants.gov/v1"

// HandleGrantsSearch searches grant opportunities.
func HandleGrantsSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q", "keywords")
	if query == "" {
		return err("query is required")
	}
	limit := getInt(args, "limit")
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Get(fmt.Sprintf(
		"%s/opportunities/search?q=%s&limit=%d",
		grantsAPIBase, url.QueryEscape(query), limit))
	if e != nil {
		return err(fmt.Sprintf("search failed: %v", e))
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(string(data))
}

// HandleGrantsByAgency searches grants by agency.
func HandleGrantsByAgency(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	agency, _ := getString(args, "agency", "agency_name")
	if agency == "" {
		return err("agency is required")
	}
	limit := getInt(args, "limit")
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Get(fmt.Sprintf(
		"%s/opportunities/search?agency=%s&limit=%d",
		grantsAPIBase, url.QueryEscape(agency), limit))
	if e != nil {
		return err(fmt.Sprintf("search by agency failed: %v", e))
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(string(data))
}

// HandleGrantsByCategory searches grants by funding category.
func HandleGrantsByCategory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	category, _ := getString(args, "category", "cat")
	if category == "" {
		return err("category is required")
	}
	limit := getInt(args, "limit")
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Get(fmt.Sprintf(
		"%s/opportunities/search?fundingCategory=%s&limit=%d",
		grantsAPIBase, url.QueryEscape(category), limit))
	if e != nil {
		return err(fmt.Sprintf("search by category failed: %v", e))
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(string(data))
}

// HandleGrantsTrends returns funding trend analysis.
func HandleGrantsTrends(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	agency, _ := getString(args, "agency", "agency_name")
	years := getInt(args, "years")
	if years <= 0 || years > 10 {
		years = 3
	}

	path := fmt.Sprintf("%s/opportunities/trends?years=%d", grantsAPIBase, years)
	if agency != "" {
		path += "&agency=" + url.QueryEscape(agency)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Get(path)
	if e != nil {
		return err(fmt.Sprintf("trends request failed: %v", e))
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(string(data))
}
