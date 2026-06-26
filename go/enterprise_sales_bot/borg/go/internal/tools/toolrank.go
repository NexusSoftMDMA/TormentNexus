package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client is a shared HTTP client with a 30-second timeout.
var Client = http.DefaultClient

// HandleSearchRank implements a tool to search for a term and return a mock ranking result.
// It simulates fetching data from an external source.
func HandleSearchRank(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 10
	}

	// Simulate a search request construction
	params := url.Values{}
	params.Set("q", query)
	params.Set("limit", strconv.Itoa(limit))

	// In a real scenario, this would hit an API. Here we simulate a successful response.
	// We use a dummy URL to demonstrate the pattern without external dependencies.
	targetURL := "https://example.com/search?" + params.Encode()

	req, reqErr := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := Client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch data: %v", fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response body: %v", readErr))
}

	// Simulate parsing logic
	result := fmt.Sprintf("Search results for '%s' (limit %d): %s", query, limit, string(body))
	return ok(result)
}

// HandleGetStats implements a tool to retrieve mock statistics.
func HandleGetStats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	category, _ :=getString(args, "category")
	if category == "" {
		category = "general"
	}

	// Simulate fetching stats
	stats := map[string]interface{}{
		"category": category,
		"total":    100,
		"active":   42,
		"updated":  time.Now().Format(time.RFC3339),
	}

	jsonData, marshalErr := json.Marshal(stats)
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to marshal stats: %v", marshalErr))
}

	return ok(string(jsonData))
}

// HandleCheckStatus implements a tool to check the status of a specific item.
func HandleCheckStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	itemID, _ :=getString(args, "id")
	if itemID == "" {
		return err("id parameter is required")
}

	verbose, _ :=getBool(args, "verbose")

	status := "active"
	if strings.HasPrefix(itemID, "inactive_") {
		status = "inactive"
	}

	response := fmt.Sprintf("Item %s is %s", itemID, status)
	if verbose {
		response += " (detailed mode)"
	}

	return ok(response)
}

// HandleListItems implements a tool to list items with optional sorting.
func HandleListItems(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sortOrder, _ :=getString(args, "sort")
	if sortOrder == "" {
		sortOrder = "asc"
	}

	items := []string{"item-a", "item-b", "item-c", "item-d"}

	if sortOrder == "desc" {
		// Simple reverse for demo
		for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
			items[i], items[j] = items[j], items[i]
		}
	}

	result := strings.Join(items, ", ")
	return ok(result)
}