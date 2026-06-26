package tools

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// HandleFetchURL fetches the content of a URL and returns the response body.
func HandleFetchURL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ :=getString(args, "url")
	if urlStr == "" {
		return err("Missing required argument: url")
}

	parsedURL, parseErr := url.Parse(urlStr)
	if parseErr != nil {
		return err("Invalid URL: " + parseErr.Error())
}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return err("URL scheme must be http or https")
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if reqErr != nil {
		return err("Failed to create request: " + reqErr.Error())
}

	req.Header.Set("User-Agent", "MCP-Fetcher/1.0")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err("Failed to fetch URL: " + fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err("Failed to read response body: " + readErr.Error())
}

	return ok(string(body))
}

// HandleFetchStatus fetches a URL and returns the HTTP status code.
func HandleFetchStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ :=getString(args, "url")
	if urlStr == "" {
		return err("Missing required argument: url")
}

	parsedURL, parseErr := url.Parse(urlStr)
	if parseErr != nil {
		return err("Invalid URL: " + parseErr.Error())
}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return err("URL scheme must be http or https")
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if reqErr != nil {
		return err("Failed to create request: " + reqErr.Error())
}

	req.Header.Set("User-Agent", "MCP-Fetcher/1.0")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err("Failed to fetch URL: " + fetchErr.Error())
}

	defer resp.Body.Close()

	statusText := "HTTP " + strconv.Itoa(resp.StatusCode)
	return ok(statusText)
}