//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func HandleTavilySearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	maxResults, _ :=getInt(args, "max_results")
	if maxResults < 1 {
		maxResults = 5
	}
	apiKey := os.Getenv("TAVILY_API_KEY")
	if apiKey == "" {
		return err("TAVILY_API_KEY not set")
}

	u, e := url.Parse("https://api.tavily.com/search")
	if e != nil {
		return err("failed to parse URL")
}

	q := u.Query()
	q.Set("api_key", apiKey)
	q.Set("query", query)
	q.Set("max_results", fmt.Sprintf("%d", maxResults))
	u.RawQuery = q.Encode()
	resp, e := http.DefaultClient.Get(u.String())
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}
