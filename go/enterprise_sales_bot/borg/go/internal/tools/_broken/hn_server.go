package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetTopStories_hn_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://hacker-news.firebaseio.com/v0/topstories.json", nil)
	if e != nil {
		return err("failed to create request")
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch top stories")
	}
	defer resp.Body.Close()
	var ids []int
	if e := json.NewDecoder(resp.Body).Decode(&ids); e != nil {
		return err("failed to decode response")
	}
	limit := 10
	if len(ids) < limit {
		limit = len(ids)

	return ok(fmt.Sprintf("Top %d story IDs: %v", limit, ids[:limit]))
}

}

func HandleGetStory_hn_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "id")
	if id == 0 {
		return err("missing id argument")
	}
	url := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", id)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch story")
	}
	defer resp.Body.Close()
	var story struct {
		Title string `json:"title"`
		By    string `json:"by"`
		URL   string `json:"url"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&story); e != nil {
		return err("failed to decode story")
	}
	return ok(fmt.Sprintf("Story: %s by %s (URL: %s)", story.Title, story.By, story.URL))
}