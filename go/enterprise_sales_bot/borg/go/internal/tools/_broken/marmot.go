package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// getMarmotConfig reads the Marmot host and API key from environment.
func getMarmotConfig() (host, apiKey string) {
	host = os.Getenv("MARMOT_HOST")
	apiKey = os.Getenv("MARMOT_API_KEY")
	if host == "" {
		host = "http://localhost:8080"
	}
	return
}

// marmotRequest performs an HTTP request to the Marmot API.
func marmotRequest(ctx context.Context, method, path string, query url.Values) ([]byte, error) {
	host, apiKey := getMarmotConfig()
	baseURL := strings.TrimRight(host, "/")

	u, parseErr := url.Parse(baseURL + path)
	if parseErr != nil {
		return nil, fmt.Errorf("invalid Marmot URL: %w", parseErr)
}

	if query != nil {
		u.RawQuery = query.Encode()

	req, reqErr := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if reqErr != nil {
		return nil, fmt.Errorf("failed to create request: %w", reqErr)
}

	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)

	req.Header.Set("Accept", "application/json")

	client := http.DefaultClient
	resp, doErr := client.Do(req)
	if doErr != nil {
		return nil, fmt.Errorf("request failed: %w", doErr)
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read response: %w", readErr)
}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
}

	return body, nil
}

}
}

// HandleDiscoverData implements the discover_data MCP tool.
// Unified data discovery supporting search, lookup by ID/MRN, and filtering.
func HandleDiscoverData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Determine mode: direct lookup by id or mrn vs. search
	id, _ :=getString(args, "id")
	mrn, _ :=getString(args, "mrn")
	queryStr, _ :=getString(args, "query")

	if id != "" {
		path := "/api/v1/assets/" + url.PathEscape(id)
		body, apiErr := marmotRequest(ctx, "GET", path, nil)
		if apiErr != nil {
			return err(apiErr.Error())
}

		return ok(string(body))
}

	if mrn != "" {
		// MRNs can be looked up by encoding as the id path segment
		path := "/api/v1/assets/" + url.PathEscape(mrn)
		body, apiErr := marmotRequest(ctx, "GET", path, nil)
		if apiErr != nil {
			return err(apiErr.Error())
}

		return ok(string(body))
}

	// Build query parameters for search
	params := url.Values{}
	if queryStr != "" {
		params.Set("q", queryStr)

	if t := getString(args, "type"); t != "" {
		params.Set("type", t)

	if provider := getString(args, "provider"); provider != "" {
		params.Set("provider", provider)

	if tags := getString(args, "tags"); tags != "" {
		// Accept comma-separated tags
		params.Set("tags", tags)

	if limit := getInt(args, "limit"); limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))

	if offset := getInt(args, "offset"); offset > 0 {
		params.Set("offset", fmt.Sprintf("%d", offset))

	body, apiErr := marmotRequest(ctx, "GET", "/api/v1/assets/search", params)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(string(body))
}

}
}
}
}
}
}

// HandleFindOwnership implements the find_ownership MCP tool.
// Bidirectional ownership queries for assets and glossary terms.
func HandleFindOwnership(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	assetID, _ :=getString(args, "asset_id")
	userID, _ :=getString(args, "user_id")
	termID, _ :=getString(args, "term_id")

	if assetID != "" {
		// Fetch asset details (owners are returned in the response)
		path := "/api/v1/assets/" + url.PathEscape(assetID)
		body, apiErr := marmotRequest(ctx, "GET", path, nil)
		if apiErr != nil {
			return err(apiErr.Error())
}

		return ok(string(body))
}

	if termID != "" {
		// Fetch glossary term (may contain owner info)
		path := "/api/v1/glossary/" + url.PathEscape(termID)
		body, apiErr := marmotRequest(ctx, "GET", path, nil)
		if apiErr != nil {
			return err(apiErr.Error())
}

		return ok(string(body))
}

	if userID != "" {
		// Fallback: search assets with a user filter (if supported)
		params := url.Values{}
		params.Set("q", "owner:"+userID) // heuristic, may vary by Marmot version
		body, apiErr := marmotRequest(ctx, "GET", "/api/v1/assets/search", params)
		if apiErr != nil {
			return err(apiErr.Error())
}

		return ok(string(body))
}

	return err("one of 'asset_id', 'user_id', or 'term_id' is required")
}

// HandleLookupTerm implements the lookup_term MCP tool.
// Business glossary lookups by ID or search query.
func HandleLookupTerm(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	queryStr, _ :=getString(args, "query")

	if id != "" {
		path := "/api/v1/glossary/" + url.PathEscape(id)
		body, apiErr := marmotRequest(ctx, "GET", path, nil)
		if apiErr != nil {
			return err(apiErr.Error())
}

		return ok(string(body))
}

	if queryStr != "" {
		params := url.Values{}
		params.Set("q", queryStr)
		body, apiErr := marmotRequest(ctx, "GET", "/api/v1/glossary/search", params)
		if apiErr != nil {
			return err(apiErr.Error())
}

		return ok(string(body))
}

	return err("either 'id' or 'query' is required")
}