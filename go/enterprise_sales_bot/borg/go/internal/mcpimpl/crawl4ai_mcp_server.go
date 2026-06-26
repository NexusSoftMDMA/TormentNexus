package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func HandleScrape_crawl4ai_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "url")
	if target == "" {
		return err("missing url parameter")
}

	apiURL := fmt.Sprintf("http://localhost:8000/scrape?url=%s", url.QueryEscape(target))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err("scrape request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("read response failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON response: " + e.Error())
}

	return ok(fmt.Sprintf("Scraped %s successfully", target))
}

func HandleCrawl_crawl4ai_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	startURL, _ :=getString(args, "start_url")
	if startURL == "" {
		return err("missing start_url parameter")
}

	maxPages, _ :=getInt(args, "max_pages")
	if maxPages < 1 {
		maxPages = 10
	}
	apiURL := fmt.Sprintf("http://localhost:8000/crawl?start_url=%s&max_pages=%d", url.QueryEscape(startURL), maxPages)
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err("crawl request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("read response failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON response: " + e.Error())
}

	return success("Crawl completed")
}