package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleRedditOptionsTraderRot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	subreddit, _ :=getString(args, "subreddit")
	if subreddit == "" {
		subreddit = "options"
	}
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	url := fmt.Sprintf("https://www.reddit.com/r/%s/hot.json?limit=%d", subreddit, limit)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("User-Agent", "MCP-Client/0.1")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data struct {
		Data struct {
			Children []struct {
				Data struct {
					Title string `json:"title"`
				} `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	var result string
	for i, child := range data.Data.Children {
		result += fmt.Sprintf("%d. %s\n", i+1, child.Data.Title)

	if result == "" {
		result = "No posts found."
	}
	return ok(result)
}
}