package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// HandleGetTopStories fetches top stories from Hacker News
func HandleGetTopStories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://hacker-news.firebaseio.com/v0/topstories.json")
	if e != nil {
		return err("Failed to fetch top stories: " + e.Error())
}

	defer resp.Body.Close()

	var ids []int
	if e := json.NewDecoder(resp.Body).Decode(&ids); e != nil {
		return err("Failed to decode top stories: " + e.Error())
}

	limit := 10
	if len(ids) < limit {
		limit = len(ids)

	var stories []string
	for i := 0; i < limit; i++ {
		storyID := ids[i]
		itemURL := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", storyID)
		resp2, e := http.DefaultClient.Get(itemURL)
		if e != nil {
			continue
		}
		var item struct {
			Title string `json:"title"`
			URL   string `json:"url"`
		}
		if e := json.NewDecoder(resp2.Body).Decode(&item); e != nil {
			resp2.Body.Close()
			continue
		}
		resp2.Body.Close()
		if item.Title != "" {
			stories = append(stories, fmt.Sprintf("- %s (%s)", item.Title, item.URL))

	}

	result := fmt.Sprintf("Top %d stories:\n", limit)
	for _, s := range stories {
		result += s + "\n"
	}
	return ok(result)
}
}
}