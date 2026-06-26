package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleCrawl_crawlbase_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing url")
}

	token, _ :=getString(args, "token")
	if token == "" {
		return err("missing token")
}

	apiURL := fmt.Sprintf("https://api.crawlbase.com/?token=%s&url=%s", token, url)
	req, e := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if e != nil {
		return err(fmt.Sprintf("request creation failed: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json decode failed: %v", e))
}

	return ok(fmt.Sprintf("Crawl result: %s", string(body)))
}

func HandleScrape_crawlbase_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing url")
}

	selector, _ :=getString(args, "selector")
	if selector == "" {
		return err("missing selector")
}

	token, _ :=getString(args, "token")
	if token == "" {
		return err("missing token")
}

	apiURL := fmt.Sprintf("https://api.crawlbase.com/scraper?token=%s&url=%s&selector=%s", token, url, selector)
	req, e := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if e != nil {
		return err(fmt.Sprintf("request creation failed: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned %d: %s", resp.StatusCode, string(body)))
}

	// Remove the trailing newline etc.
	trimmed := strings.TrimSpace(string(body))
	return ok(fmt.Sprintf("Scraped content: %s", trimmed))
}