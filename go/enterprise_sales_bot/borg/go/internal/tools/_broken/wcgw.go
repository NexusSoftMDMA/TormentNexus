package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetTopStories_wcgw(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit == 0 || limit > 30 {
		limit = 10
	}
	resp, e := http.DefaultClient.Get("https://hacker-news.firebaseio.com/v0/topstories.json")
	if e != nil {
		return err("failed to fetch top stories: " + e.Error())
}

	defer resp.Body.Close()
	var ids []int
	if e := json.NewDecoder(resp.Body).Decode(&ids); e != nil {
		return err("failed to decode: " + e.Error())
}

	if limit > len(ids) {
		limit = len(ids)

	ids = ids[:limit]
	strs := make([]string, limit)
	for i, id := range ids {
		strs[i] = fmt.Sprintf("%d", id)

	return ok(fmt.Sprintf("Top %d story IDs: %v", limit, strs))
}

}
}

func HandleGetStory_wcgw(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "id")
	if id == 0 {
		return err("id is required")
}

	url := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch story: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode: " + e.Error())
}

	title, found := data["title"].(string)
	if !found {
		return err("story not found")
}

	return ok(fmt.Sprintf("Story %d: %s", id, title))
}