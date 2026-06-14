package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetTopStories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://hacker-news.firebaseio.com/v0/topstories.json")
	if e != nil {
		return err("Failed to fetch top stories: " + e.Error())
}

	defer resp.Body.Close()
	var ids []int
	if e := json.NewDecoder(resp.Body).Decode(&ids); e != nil {
		return err("Failed to decode response: " + e.Error())
}

	if len(ids) > 10 {
		ids = ids[:10]
	}
	return success(fmt.Sprintf("Top story IDs: %v", ids))
}

func HandleGetStory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "id")
	if id == 0 {
		return err("Missing or invalid 'id' parameter")
}

	url := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("Failed to fetch story: " + e.Error())
}

	defer resp.Body.Close()
	var story map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&story); e != nil {
		return err("Failed to decode story: " + e.Error())
}

	if title, found := story["title"].(string); found {
		return success(fmt.Sprintf("Story: %s", title))
}

	return success("Story data: " + fmt.Sprintf("%v", story))
}