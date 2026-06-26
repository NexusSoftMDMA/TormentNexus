package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

func HandleSearchArticles_nyt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	page, _ :=getInt(args, "page")
	apiKey := os.Getenv("NYT_API_KEY")
	url := fmt.Sprintf("https://api.nytimes.com/svc/search/v2/articlesearch.json?q=%s&page=%d&api-key=%s", query, page, apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch articles: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	response := result["response"].(map[string]interface{})
	docs := response["docs"].([]interface{})
	return ok(fmt.Sprintf("Found %d articles", len(docs)))
}

func HandleTopStories_nyt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	section, _ :=getString(args, "section")
	if section == "" {
		section = "home"
	}
	apiKey := os.Getenv("NYT_API_KEY")
	url := fmt.Sprintf("https://api.nytimes.com/svc/topstories/v2/%s.json?api-key=%s", section, apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch top stories: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	results := result["results"].([]interface{})
	return ok(fmt.Sprintf("Found %d top stories for %s", len(results), section))
}