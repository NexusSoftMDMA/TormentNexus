package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetStory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "id")
	if id == 0 {
		return err("story id required")
}

	url := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch story: " + e.Error())
}

	defer resp.Body.Close()
	var story map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&story); e != nil {
		return err("failed to decode story: " + e.Error())
}

	if story == nil {
		return err("story not found")
}

	return ok(story)
}

func HandleGetTopStories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://hacker-news.firebaseio.com/v0/topstories.json")
	if e != nil {
		return err("failed to fetch top stories: " + e.Error())
}

	defer resp.Body.Close()
	var ids []int
	if e := json.NewDecoder(resp.Body).Decode(&ids); e != nil {
		return err("failed to decode top stories: " + e.Error())
}

	limit, _ :=getInt(args, "limit")
	if limit <= 0 || limit > len(ids) {
		limit = 10
	}
	return ok(ids[:limit])
}