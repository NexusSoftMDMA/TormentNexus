package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

func HandleScrape(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ :=getString(args, "url")
	onlyMainContent, _ :=getBool(args, "only_main_content")
	if urlStr == "" {
		return err("url is required")
}

	u, e := url.Parse(urlStr)
	if e != nil {
		return err(fmt.Sprintf("invalid URL: %v", e))
}

	if !isSafeURL(u) {
		return err("unsafe URL")
}

	client := http.DefaultClient
	resp, fetchErr := client.Get(urlStr)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch URL: %v", fetchErr))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("HTTP error: %s", resp.Status))
}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return err("content is not HTML")
}

	var content string
	if onlyMainContent {
		content = extractMainContent(resp.Body)
	} else {
		content = extractFullContent(resp.Body)

	if content == "" {
		return err("no content extracted")
}

	return ok(content)
}

}

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	numResults, _ :=getInt(args, "num_results")
	scrape, _ :=getBool(args, "scrape")
	if query == "" {
		return err("query is required")
}

	if numResults <= 0 {
		numResults = 5
	}
	apiKey := os.Getenv("SERPER_API_KEY")
	if apiKey == "" {
		return err("SERPER_API_KEY environment variable not set")
}

	searchURL := fmt.Sprintf("https://api.serper.dev/search?api_key=%s&q=%s&num=%d", apiKey, url.QueryEscape(query), numResults)
	client := http.DefaultClient
	resp, fetchErr := client.Get(searchURL)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to search: %v", fetchErr))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("search API error: %s", resp.Status))
}

	var result struct {
		Results []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Description string `json:"description"`
		} `json:"organic"`
	}
	if parseErr := json.NewDecoder(resp.Body).Decode(&result); parseErr != nil {
		return err(fmt.Sprintf("failed to parse search results: %v", parseErr))
}

	if len(result.Results) == 0 {
		return err("no search results found")
}

	if !scrape {
		var output strings.Builder
		for i, item := range result.Results {
			fmt.Fprintf(&output, "%d. %s\n%s\n%s\n\n", i+1, item.Title, item.URL, item.Description)

		return ok(output.String())
	}
	var scrapedResults []string
	for _, item := range result.Results {
		if !isSafeURLString(item.URL) {
			continue
		}
		resp, fetchErr := client.Get(item.URL)
		if fetchErr != nil {
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			continue
		}
		content, e := io.ReadAll(resp.Body)
		if e != nil {
			continue
		}
		extractedContent := extractMainContent(io.NopCloser(strings.NewReader(string(content))))
		if extractedContent != "" {
			scrapedResults = append(scrapedResults, fmt.Sprintf("Title: %s\nURL: %s\nContent:\n%s\n", item.Title, item.URL, extractedContent))

	}
	if len(scrapedResults) == 0 {
		return err("no content scraped from search results")
}

	return ok(strings.Join(scrapedResults, "\n---\n"))
}

}
}

func HandleCrawl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	startURL, _ :=getString(args, "start_url")
	depth, _ :=getInt(args, "depth")
	maxPages, _ :=getInt(args, "max_pages")
	concurrency, _ :=getInt(args, "concurrency")
	useSitemap, _ :=getBool(args, "use_sitemap")
	if startURL == "" {
		return err("start_url is required")
}

	if depth <= 0 {
		depth = 1
	}
	if maxPages <= 0 {
		maxPages = 10
	}
	if concurrency <= 0 {
		concurrency = 2
	}
	u, e := url.Parse(startURL)
	if e != nil {
		return err(fmt.Sprintf("invalid start URL: %v", e))
}

	if !isSafeURL(u) {
		return err("unsafe start URL")
}

	client := http.DefaultClient
	var urls []string
	if useSitemap {
		sitemapURL := fmt.Sprintf("%s://%s/sitemap.xml", u.Scheme, u.Host)
		resp, fetchErr := client.Get(sitemapURL)
		if fetchErr != nil {
			return err(fmt.Sprintf("failed to fetch sitemap: %v", fetchErr))
}

		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			sitemapContent, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				return err(fmt.Sprintf("failed to read sitemap: %v", readErr))
}

			urls = extractURLsFromSitemap(string(sitemapContent))

	}
	if len(urls) == 0 {
		urls = append(urls, startURL)

	visited := make(map[string]bool)
	var results []string
	for len(urls) > 0 && len(results) < maxPages {
		currentURL := urls[0]
		urls = urls[1:]
		if visited[currentURL] {
			continue
		}
		visited[currentURL] = true
		if !isSafeURLString(currentURL) {
			continue
		}
		resp, fetchErr := client.Get(currentURL)
		if fetchErr != nil {
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			continue
		}
		content, e := io.ReadAll(resp.Body)
		if e != nil {
			continue
		}
		extractedContent := extractMainContent(io.NopCloser(strings.NewReader(string(content))))
		if extractedContent != "" {
			results = append(results, fmt.Sprintf("URL: %s\nContent:\n%s\n", currentURL, extractedContent))

		if len(results) >= maxPages {
			break
		}
		if depth > 1 {
			links := extractLinks(io.NopCloser(strings.NewReader(string(content))))
			for _, link := range links {
				if isSafeURLString(link) && !visited[link] {
					urls = append(urls, link)

			}
		}
	}
	if len(results) == 0 {
		return err("no content crawled")
}

	return ok(strings.Join(results, "\n---\n"))
}

}
}
}
}

func isSafeURL(u *url.URL) bool {
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	host := u.Hostname()
	if host == "" {
		return false
	}
	if strings.HasPrefix(host, "localhost") || strings.HasPrefix(host, "127.0.0.1") || strings.HasPrefix(host, "::1") {
		return false
	}
	return true
}

func isSafeURLString(urlStr string) bool {
	u, e := url.Parse(urlStr)
	if e != nil {
		return false
	}
	return isSafeURL(u)
}

func extractMainContent(body io.ReadCloser) string {
	content, e := io.ReadAll(body)
	if e != nil {
		return ""
	}
	return string(content)
}

func extractFullContent(body io.ReadCloser) string {
	content, e := io.ReadAll(body)
	if e != nil {
		return ""
	}
	return string(content)
}

func extractURLsFromSitemap(sitemap string) []string {
	re := regexp.MustCompile(`<loc>(.*?)</loc>`)
	matches := re.FindAllStringSubmatch(sitemap, -1)
	var urls []string
	for _, match := range matches {
		if len(match) > 1 {
			urls = append(urls, match[1])

	}
	return urls
}

}

func extractLinks(body io.ReadCloser) []string {
	content, e := io.ReadAll(body)
	if e != nil {
		return nil
	}
	re := regexp.MustCompile(`href=["'](.*?)["']`)
	matches := re.FindAllStringSubmatch(string(content), -1)
	var links []string
	for _, match := range matches {
		if len(match) > 1 {
			links = append(links, match[1])

	}
	return links
}
}