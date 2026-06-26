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
	"strconv"
	"strings"
	"time"
)

// shared HTTP client with timeout
var http.DefaultClient = http.DefaultClient

// HandleFetch fetches the content of a URL and returns it as plain text.
func HandleFetch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ :=getString(args, "url")
	if urlStr == "" {
		return err("missing url argument")
}

	// validate URL
	parsed, parseErr := url.ParseRequestURI(urlStr)
	if parseErr != nil {
		return err(fmt.Sprintf("invalid url: %v", parseErr))
}

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	return ok(string(body))
}

// HandleExtractLinks fetches a page and extracts all href links.
// Returns a JSON array of strings.
func HandleExtractLinks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ :=getString(args, "url")
	if urlStr == "" {
		return err("missing url argument")
}

	parsed, parseErr := url.ParseRequestURI(urlStr)
	if parseErr != nil {
		return err(fmt.Sprintf("invalid url: %v", parseErr))
}

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	content := string(body)

	// simple regex to find href attributes
	hrefRe := regexp.MustCompile(`(?i)href\s*=\s*["']([^"']+)["']`)")
	matches := hrefRe.FindAllStringSubmatch(content, -1)

	linksMap := make(map[string]struct{})
	for _, m := range matches {
		if len(m) > 1 {
			linksMap[m[1]] = struct{}{}
		}
	}

	links := make([]string, 0, len(linksMap))
	for l := range linksMap {
		links = append(links, l)

	sort.Strings(links)

	jsonBytes, jsonErr := json.Marshal(links)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(jsonBytes))
}

}

// HandleDownload downloads a file from a URL to a destination path.
// If the destination directory does not exist, it is created.
func HandleDownload(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ :=getString(args, "url")
	destPath, _ :=getString(args, "dest")
	if urlStr == "" || destPath == "" {
		return err("missing url or dest argument")
}

	parsed, parseErr := url.ParseRequestURI(urlStr)
	if parseErr != nil {
		return err(fmt.Sprintf("invalid url: %v", parseErr))
}

	// Ensure destination directory exists
	dir := filepath.Dir(destPath)
	if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
		return err(mkErr.Error())
}

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
}

	outFile, fileErr := os.Create(destPath)
	if fileErr != nil {
		return err(fileErr.Error())
}

	defer outFile.Close()

	_, copyErr := io.Copy(outFile, resp.Body)
	if copyErr != nil {
		return err(copyErr.Error())
}

	return ok(fmt.Sprintf("downloaded to %s", destPath))
}

// HandleOpenUrl is a placeholder that simply acknowledges the request.
// In a real environment it could launch the default browser.
func HandleOpenUrl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ :=getString(args, "url")
	if urlStr == "" {
		return err("missing url argument")
}

	// No actual opening performed; just confirm receipt.
	return ok(fmt.Sprintf("opened url: %s", urlStr))
}