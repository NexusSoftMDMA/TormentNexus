package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	versionRegex = regexp.MustCompile(`v\d+\.\d+\.\d+`)
)

func HandleVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Get the latest version from GitHub releases
	client := http.DefaultClient
	resp, reqErr := client.Get("https://api.github.com/repos/asciimoo/hister/releases/latest")
	if reqErr != nil {
		return err(fmt.Sprintf("Failed to fetch latest version: %v", reqErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("GitHub API returned status: %s", resp.Status))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("Failed to read response body: %v", readErr))
}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if parseErr := json.Unmarshal(body, &release); parseErr != nil {
		return err(fmt.Sprintf("Failed to parse release data: %v", parseErr))
}

	// Extract version number from tag name
	version := versionRegex.FindString(release.TagName)
	if version == "" {
		return err("Could not extract version number from tag")
}

	return ok(fmt.Sprintf("Current Hister version: %s", version))
}

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("Query parameter is required")
}

	// Construct search URL
	searchURL := fmt.Sprintf("https://hister.org/search?q=%s", url.QueryEscape(query))

	// Make the search request
	client := http.DefaultClient
	resp, reqErr := client.Get(searchURL)
	if reqErr != nil {
		return err(fmt.Sprintf("Failed to perform search: %v", reqErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Search returned status: %s", resp.Status))
}

	// Read and return the response
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("Failed to read search results: %v", readErr))
}

	return ok(string(body))
}

func HandleIndex(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("URL parameter is required")
}

	// Validate URL format
	if _, parseErr := url.ParseRequestURI(url); parseErr != nil {
		return err(fmt.Sprintf("Invalid URL format: %v", parseErr))
}

	// Construct index URL
	indexURL := fmt.Sprintf("https://hister.org/index?url=%s", url.QueryEscape(url))

	// Make the index request
	client := http.DefaultClient
	resp, reqErr := client.Get(indexURL)
	if reqErr != nil {
		return err(fmt.Sprintf("Failed to index URL: %v", reqErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Indexing returned status: %s", resp.Status))
}

	// Read and return the response
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("Failed to read index response: %v", readErr))
}

	return ok(fmt.Sprintf("Successfully indexed %s: %s", url, string(body)))
}

func HandleListExtractors(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Get the list of available extractors
	client := http.DefaultClient
	resp, reqErr := client.Get("https://hister.org/api/extractors")
	if reqErr != nil {
		return err(fmt.Sprintf("Failed to fetch extractors: %v", reqErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status: %s", resp.Status))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("Failed to read extractors data: %v", readErr))
}

	var extractors []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Enabled     bool   `json:"enabled"`
	}

	if parseErr := json.Unmarshal(body, &extractors); parseErr != nil {
		return err(fmt.Sprintf("Failed to parse extractors data: %v", parseErr))
}

	// Sort extractors by name
	sort.Slice(extractors, func(i, j int) bool {
		return strings.Compare(extractors[i].Name, extractors[j].Name) < 0
	})

	// Format the response
	var response strings.Builder
	response.WriteString("Available extractors:\n")
	for _, e := range extractors {
		response.WriteString(fmt.Sprintf("- %s: %s (Enabled: %t)\n", e.Name, e.Description, e.Enabled))

	return ok(response.String())
}

}

func HandlePreview(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	docID, _ :=getString(args, "id")
	if docID == "" {
		return err("Document ID parameter is required")
}

	// Construct preview URL
	previewURL := fmt.Sprintf("https://hister.org/preview/%s", url.QueryEscape(docID))

	// Make the preview request
	client := http.DefaultClient
	resp, reqErr := client.Get(previewURL)
	if reqErr != nil {
		return err(fmt.Sprintf("Failed to fetch preview: %v", reqErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Preview returned status: %s", resp.Status))
}

	// Read and return the response
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("Failed to read preview data: %v", readErr))
}

	return ok(string(body))
}

func HandleDelete(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	docID, _ :=getString(args, "id")
	if docID == "" {
		return err("Document ID parameter is required")
}

	// Construct delete URL
	deleteURL := fmt.Sprintf("https://hister.org/api/documents/%s", url.QueryEscape(docID))

	// Make the delete request
	client := http.DefaultClient
	req, reqErr := http.NewRequest(http.MethodDelete, deleteURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("Failed to create delete request: %v", reqErr))
}

	resp, reqErr := client.Do(req)
	if reqErr != nil {
		return err(fmt.Sprintf("Failed to delete document: %v", reqErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Delete returned status: %s", resp.Status))
}

	// Read and return the response
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("Failed to read delete response: %v", readErr))
}

	return ok(fmt.Sprintf("Successfully deleted document %s: %s", docID, string(body)))
}