package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchReddit_reddit_research_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	limit, _ :=getInt(args, "limit")
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	u := fmt.Sprintf("https://www.reddit.com/search.json?q=%s&limit=%d", url.QueryEscape(query), limit)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var data struct {
		Data struct {
			Children []struct {
				Data struct {
					Title string `json:"title"`
					URL   string `json:"url"`
				} `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}
	if e = json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON")
}

	var results []string
	for _, child := range data.Data.Children {
		results = append(results, fmt.Sprintf("- %s (%s)", child.Data.Title, child.Data.URL))

	if len(results) == 0 {
		return ok("No results found.")
}

	out := "Reddit search results:\n"
	for _, r := range results {
		out += r + "\n"
	}
	return success(out)
}

}

func HandleGetSubredditInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sub, _ :=getString(args, "subreddit")
	if sub == "" {
		return err("subreddit is required")
}

	u := fmt.Sprintf("https://www.reddit.com/r/%s/about.json", url.PathEscape(sub))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var data struct {
		Data struct {
			DisplayName string `json:"display_name"`
			Description string `json:"public_description"`
			Subscribers int    `json:"subscribers"`
		} `json:"data"`
	}
	if e = json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON")
}

	out := fmt.Sprintf("Subreddit: %s\nDescription: %s\nSubscribers: %d", data.Data.DisplayName, data.Data.Description, data.Data.Subscribers)
	return success(out)
}