package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetStory_kotlin_sdk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
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
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	data, e := json.Marshal(result)
	if e != nil {
		return err("failed to marshal result: " + e.Error())
}

	return success(string(data))
}

func HandleGetTopStories_kotlin_sdk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	url := "https://hacker-news.firebaseio.com/v0/topstories.json"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch top stories: " + e.Error())
}

	defer resp.Body.Close()
	var ids []int
	if e := json.NewDecoder(resp.Body).Decode(&ids); e != nil {
		return err("failed to decode top stories: " + e.Error())
}

	if len(ids) > limit {
		ids = ids[:limit]
	}
	data, e := json.Marshal(ids)
	if e != nil {
		return err("failed to marshal ids: " + e.Error())
}

	return success(string(data))
}