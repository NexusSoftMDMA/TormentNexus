package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// http.DefaultClient is a shared client with a timeout.
var http.DefaultClient = http.DefaultClient

// fetchReadme retrieves the README.md content from the repository.
func fetchReadme() (string, error) {
	const readmeURL = "https://raw.githubusercontent.com/taielab/awesome-hacking-lists/master/README.md"
	resp, fetchErr := http.DefaultClient.Get(readmeURL)
	if fetchErr != nil {
		return "", fmt.Errorf("failed to fetch README: %w", fetchErr)
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d while fetching README", resp.StatusCode)
}

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", fmt.Errorf("failed to read README body: %w", readErr)
}

	return string(bodyBytes), nil
}

// fetchRawFile retrieves a raw file from the repository given its relative path.
func fetchRawFile(path string) (string, error) {
	base := "https://raw.githubusercontent.com/taielab/awesome-hacking-lists/master/"
	u, parseErr := url.Parse(base + path)
	if parseErr != nil {
		return "", fmt.Errorf("invalid file path: %w", parseErr)
}

	resp, fetchErr := http.DefaultClient.Get(u.String())
	if fetchErr != nil {
		return "", fmt.Errorf("failed to fetch file: %w", fetchErr)
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d while fetching file", resp.StatusCode)
}

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", fmt.Errorf("failed to read file body: %w", readErr)
}

	return string(bodyBytes), nil
}

// HandleListResources returns a plain‑text list of all markdown links found in the README.
func HandleListResources(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	readme, fetchErr := fetchReadme()
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	lines := strings.Split(readme, "\n")
	var builder strings.Builder
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		// Markdown link pattern: - [Title](URL)
		if strings.HasPrefix(trim, "- [") && strings.Contains(trim, "](") && strings.HasSuffix(trim, ")") {
			builder.WriteString(trim)
			builder.WriteString("\n")

	}
	if builder.Len() == 0 {
		return ok("No resources found.")
}

	return ok(builder.String())
}

}

// HandleSearchResources searches the README for lines containing the provided query (case‑insensitive).
func HandleSearchResources(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	readme, fetchErr := fetchReadme()
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	lowerQuery := strings.ToLower(query)
	lines := strings.Split(readme, "\n")
	var builder strings.Builder
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), lowerQuery) {
			builder.WriteString(line)
			builder.WriteString("\n")

	}
	if builder.Len() == 0 {
		return ok(fmt.Sprintf("No matches found for \"%s\".", query))
}

	return ok(builder.String())
}

}

// HandleGetRawFile returns the raw content of a file inside the repository.
// Expected argument: "path" – the relative path to the file (e.g., "lists/cryptography.md").
func HandleGetRawFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path parameter is required")
}

	content, fetchErr := fetchRawFile(path)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	return ok(content)
}